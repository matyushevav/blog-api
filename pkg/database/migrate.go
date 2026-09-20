package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
)

// migrationsDir - папка с SQL-миграциями.
const migrationsDir = "migrations"

// Migrate выполняет все SQL-миграции из папки migrations в порядке имен файлов.
func Migrate(db *sql.DB) error {
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".sql" {
			continue
		}

		files = append(files, entry.Name())
	}

	// Сортируем файлы
	sort.Strings(files)

	for _, name := range files {
		log.Printf("Running migration: %s", name)

		content, err := os.ReadFile(filepath.Join(migrationsDir, name))
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", name, err)
		}

		if _, err := db.Exec(string(content)); err != nil {
			log.Printf("Migration failed: %s: %v", name, err)

			return fmt.Errorf("failed to run migration %s: %w", name, err)
		}

		log.Printf("Successfully applied migration: %s", name)
	}

	return nil
}
