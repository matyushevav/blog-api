package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"advanced-blog-management-system/internal/handler"
	"advanced-blog-management-system/internal/logger"
	"advanced-blog-management-system/internal/middleware"
	"advanced-blog-management-system/internal/repository"
	"advanced-blog-management-system/internal/service"
	"advanced-blog-management-system/pkg/auth"
	"advanced-blog-management-system/pkg/database"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
)

func main() {
	// Ищем папку migrations в текущей директории и выше (до 2 уровней)
	rootDir, err := findProjectRoot()
	if err != nil {
		log.Fatalf("Failed to find project root: %v", err)
	}

	if err := os.Chdir(rootDir); err != nil {
		log.Fatalf("Failed to change directory to %s: %v", rootDir, err)
	}

	// Переносим значения из .env в переменные окружения. Если файла нет - это не ошибка
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found, using environment variables: %v", err)
	}

	// Получаем конфиги подключения БД
	dbPort, err := strconv.Atoi(getEnv("DB_PORT", "5432"))
	if err != nil {
		log.Fatalf("Invalid DB_PORT: %v", err)
	}

	dbCfg := database.Config{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     dbPort,
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", "postgres"),
		DBName:   getEnv("DB_NAME", "blog_db"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}

	// Подключаемся к БД
	db, err := database.NewPostgresDB(dbCfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close(db)

	// Запускаем миграцию данных
	if err := database.Migrate(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Инициализируем JWTManager
	// Секрет обязателен: с пустым ключом любой мог бы подписать себе токен
	jwtSecret := getEnv("JWT_SECRET", "")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET is not set")
	}

	jwtExpiryHours, err := strconv.Atoi(getEnv("JWT_EXPIRY_HOURS", "24"))
	if err != nil {
		log.Fatalf("Invalid JWT_EXPIRY_HOURS: %v", err)
	}

	jwtManager := auth.NewJWTManager(jwtSecret, jwtExpiryHours)

	// Инициализируем репозитории
	userRepo := repository.NewUserRepo(db)
	postRepo := repository.NewPostRepo(db)
	commentRepo := repository.NewCommentRepo(db)

	// Инициализируем сервисы
	userService := service.NewUserService(userRepo, jwtManager)
	postService := service.NewPostService(postRepo, userRepo)
	commentService := service.NewCommentService(commentRepo, postRepo)

	// Включаем журнал событий
	eventLogger := logger.NewEventLogger(getEnv("LOGS_FILE", "./log.txt"))
	eventLogger.Start()

	// Инициализируем handlers
	authHandler := handler.NewAuthHandler(userService)
	postHandler := handler.NewPostHandler(postService, eventLogger)
	commentHandler := handler.NewCommentHandler(commentService, eventLogger)

	// Инициализируем middleware
	loggingMW := middleware.NewLoggingMiddleware(log.New(os.Stdout, "[HTTP] ", log.LstdFlags))
	authMW := middleware.NewAuthMiddleware(jwtManager)

	// Настраиваем роутер
	router := setupRouter(authHandler, postHandler, commentHandler, loggingMW, authMW)

	// Запускаем сервер
	addr := net.JoinHostPort(getEnv("SERVER_HOST", "0.0.0.0"), getEnv("SERVER_PORT", "8080"))

	server := &http.Server{
		Addr:              addr,
		Handler:           router,
		ReadHeaderTimeout: readHeaderTimeout,
	}

	go func() {
		log.Printf("Server listening on %s", addr)

		// ErrServerClosed - штатный результат Shutdown, а не ошибка
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Установить обработчик сигналов (SIGINT, SIGTERM)
	// Блокировать main, пока не получен сигнал
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	sig := <-quit
	log.Printf("Received signal %s, shutting down...", sig)

	// Завершить сервер с таймаутом 30 секунд и закрыть БД
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	// Shutdown перестает принимать новые запросы и ждет завершения текущих
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	// Останавливаем логгер
	eventLogger.Stop()

	// БД закроет defer database.Close(db) при выходе из main
	log.Println("Server stopped")
}

const (
	// shutdownTimeout - сколько ждать завершения активных запросов при остановке
	shutdownTimeout = 30 * time.Second

	// readHeaderTimeout - защита от клиентов, которые бесконечно медленно шлют заголовки
	readHeaderTimeout = 5 * time.Second
)

// maxRootSearchDepth - на сколько уровней выше текущей директории искать корень проекта
const maxRootSearchDepth = 2

// findProjectRoot ищет директорию с папкой migrations: сначала текущую, затем до maxRootSearchDepth уровней выше.
func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	start := dir

	for i := 0; i <= maxRootSearchDepth; i++ {
		info, err := os.Stat(filepath.Join(dir, "migrations"))
		if err == nil && info.IsDir() {
			return dir, nil
		}

		dir = filepath.Dir(dir)
	}

	return "", fmt.Errorf("migrations directory not found in %s or %d levels above", start, maxRootSearchDepth)
}

// getEnv возвращает значение переменной окружения или defaultValue, если она не задана.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return defaultValue
}

// setupRouter регистрирует middleware и все маршруты API.
func setupRouter(
	authHandler *handler.AuthHandler,
	postHandler *handler.PostHandler,
	commentHandler *handler.CommentHandler,
	loggingMW *middleware.LoggingMiddleware,
	authMW *middleware.AuthMiddleware,
) *chi.Mux {
	r := chi.NewRouter()

	// Регистрируем middleware
	r.Use(loggingMW.Recovery, loggingMW.Logger, loggingMW.CORS)

	// Регистрируем публичные эндпоинты
	r.Route("/api", func(r chi.Router) {
		// POST /api/register, POST /api/login, GET /api/health
		// GET /api/posts, GET /api/posts/{id}
		// GET /api/posts/{postId}/comments
		r.Get("/health", healthHandler)
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)

		// Чтение публичное, но если токен есть, данные пользователя
		// попадут в контекст (GetByID принимает userID)
		r.Group(func(r chi.Router) {
			r.Use(middleware.ToMiddleware(authMW.OptionalAuth))

			r.Get("/posts", postHandler.GetAll)
			r.Get("/posts/{id}", postHandler.GetByID)
			r.Get("/posts/author/{authorID}", postHandler.GetByAuthor)
			r.Get("/posts/{postId}/comments", commentHandler.GetByPost)
		})

		// Регистрируем защищенные эндпоинты (требуют AuthMiddleware)
		// POST /api/posts
		// POST /api/posts/{postId}/comments
		r.Group(func(r chi.Router) {
			r.Use(middleware.ToMiddleware(authMW.RequireAuth))

			r.Get("/profile", authHandler.GetProfile)
			r.Post("/posts", postHandler.Create)
			r.Post("/posts/{postId}/comments", commentHandler.Create)
		})
	})

	return r
}

// healthHandler отвечает на проверку доступности API.
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}
