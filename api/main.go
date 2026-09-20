package main

import (
	"log"
	"os"
	"strconv"

	"advanced-blog-management-system/pkg/database"

	"github.com/joho/godotenv"
)

func main() {
	// TODO: Загрузить .env используя godotenv.Load()
	// Обработать ошибку: warning если файла нет, так как переменные могут быть в системе
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found, using environment variables: %v", err)
	}

	// TODO: Реализовать загрузку конфигурации из переменных окружения
	// Config должен содержать: ServerHost, ServerPort, DB параметры, JWT параметры
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

	// TODO: Найти корневую папку проекта (поддержка Docker)
	// Искать папку migrations в текущей директории и выше (до 2 уровней)

	// TODO: Подключиться к PostgreSQL БД
	// Использовать database.NewPostgresDB() и database.Migrate()
	db, err := database.NewPostgresDB(dbCfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close(db)

	if err := database.Migrate(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// TODO: Инициализировать JWTManager
	// auth.NewJWTManager(jwtSecret, jwtExpiryHours)

	// TODO: Инициализировать репозитории
	// NewUserRepository(db), NewPostRepository(db), NewCommentRepo(db)

	// TODO: Инициализировать сервисы
	// NewUserService(userRepo, jwtManager)
	// NewPostService(postRepo, userRepo)
	// NewCommentService(commentRepo, postRepo)

	// TODO: Инициализировать handlers
	// NewAuthHandler(userService), NewPostHandler(postService, eventLogger), NewCommentHandler(commentService, eventLogger)

	// TODO: Инициализировать middleware
	// LoggingMiddleware, AuthMiddleware

	// TODO: Настроить HTTP роутер через setupRouter()

	// TODO: Создать и запустить HTTP сервер
	// Использовать goroutine с ListenAndServe()

	// TODO: Установить обработчик сигналов (SIGINT, SIGTERM)
	// Блокировать main пока не получен сигнал

	// TODO: Graceful shutdown
	// Завершить сервер с таймаутом 30 секунд и закрыть БД

	log.Println("Server starting... (TODO: implement main.go)")
}

// getEnv возвращает значение переменной окружения или defaultValue, если она не задана.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return defaultValue
}

func setupRouter(authHandler interface{}, postHandler interface{}, commentHandler interface{}, loggingMW interface{}, authMW interface{}) interface{} {
	// TODO: Создать chi роутер

	// TODO: Зарегистрировать глобальные middleware
	// Recovery, Logger, CORS

	// TODO: Зарегистрировать публичные эндпоинты
	// POST /api/register, POST /api/login, GET /api/health
	// GET /api/posts, GET /api/posts/{id}
	// GET /api/posts/{postId}/comments

	// TODO: Зарегистрировать защищенные эндпоинты (требуют AuthMiddleware)
	// POST /api/posts
	// POST /api/posts/{postId}/comments

	// TODO: Вернуть настроенный *chi.Router

	return nil
}
