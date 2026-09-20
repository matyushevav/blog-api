package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

const (
	// maxOpenConns - максимум одновременно открытых соединений с БД.
	maxOpenConns = 25

	// maxIdleConns - максимум простаивающих соединений в пуле.
	maxIdleConns = 25

	// connMaxLifetime - время жизни одного соединения.
	connMaxLifetime = 5 * time.Minute
)

// Config содержит конфигурацию базы данных
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// NewPostgresDB открывает подключение к PostgreSQL и настраивает пул соединений.
func NewPostgresDB(cfg Config) (*sql.DB, error) {
	dsn := GetDSN(cfg)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()

		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	db.SetConnMaxLifetime(connMaxLifetime)

	log.Println("Connected to PostgreSQL database")

	return db, nil
}

// GetDSN собирает строку подключения к PostgreSQL из конфигурации.
func GetDSN(cfg Config) string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode)
}

// Close закрывает подключение к базе данных.
func Close(db *sql.DB) {
	if db == nil {
		return
	}

	if err := db.Close(); err != nil {
		log.Printf("Failed to close connection with database: %v", err)
		return
	}

	log.Println("Database connection closed")
}

// TestConnection выполняет пробный запрос и сообщает, работает ли подключение.
func TestConnection(db *sql.DB) error {
	if db == nil {
		return fmt.Errorf("connection with database is not initialized")
	}

	query := "SELECT 1"
	row := db.QueryRow(query)

	var res int
	if err := row.Scan(&res); err != nil {
		return fmt.Errorf("failed to test database connection: %w", err)
	}
	return nil
}
