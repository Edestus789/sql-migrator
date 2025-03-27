//go:build integration
// +build integration

package integration

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/Edestus789/sql-migrator/app"
	"github.com/Edestus789/sql-migrator/logger"
	"github.com/Edestus789/sql-migrator/storage"
	_ "github.com/lib/pq"
)

const (
	dbUser     = "user"
	dbPassword = "password"
	dbName     = "test_db"
	dbHost     = "localhost"
	dbPort     = "5432"
)

func TestMigrations(t *testing.T) {
	// Настройка тестовой БД
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}
	defer db.Close()

	// Инициализация мигратора
	log := logger.New()
	storage := storage.NewPostgresStorage(connStr, log)
	ctx := context.Background()

	if err := storage.Connect(ctx); err != nil {
		t.Fatalf("Failed to connect storage: %v", err)
	}
	defer storage.Close()

	// Подготовка тестовых миграций
	migrationDir := "./test_migrations"
	if err := os.MkdirAll(migrationDir, 0755); err != nil {
		t.Fatalf("Failed to create migrations dir: %v", err)
	}
	defer os.RemoveAll(migrationDir)

	// Создаем тестовые миграции
	upSQL := "CREATE TABLE IF NOT EXISTS test_table (id SERIAL PRIMARY KEY);"
	downSQL := "DROP TABLE IF EXISTS test_table;"

	upFile := fmt.Sprintf("%s/0001_test_up.sql", migrationDir)
	downFile := fmt.Sprintf("%s/0001_test_down.sql", migrationDir)

	if err := os.WriteFile(upFile, []byte(upSQL), 0644); err != nil {
		t.Fatalf("Failed to create up migration: %v", err)
	}
	if err := os.WriteFile(downFile, []byte(downSQL), 0644); err != nil {
		t.Fatalf("Failed to create down migration: %v", err)
	}

	// Тестируем миграции
	app := app.New(log, storage)

	t.Run("Apply migrations", func(t *testing.T) {
		if err := app.Up(migrationDir); err != nil {
			t.Fatalf("Up failed: %v", err)
		}

		// Проверяем, что таблица создана
		var exists bool
		err := db.QueryRow("SELECT EXISTS (SELECT FROM pg_tables WHERE tablename = 'test_table')").Scan(&exists)
		if err != nil || !exists {
			t.Fatalf("Table not created: %v", err)
		}
	})

	t.Run("Rollback migrations", func(t *testing.T) {
		if err := app.Down(migrationDir); err != nil {
			t.Fatalf("Down failed: %v", err)
		}

		// Проверяем, что таблица удалена
		var exists bool
		err := db.QueryRow("SELECT EXISTS (SELECT FROM pg_tables WHERE tablename = 'test_table')").Scan(&exists)
		if err != nil || exists {
			t.Fatalf("Table not dropped: %v", err)
		}
	})
}
