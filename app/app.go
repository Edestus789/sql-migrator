package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/Edestus789/sql-migrator/logger"
	"github.com/Edestus789/sql-migrator/processes"
	"github.com/Edestus789/sql-migrator/storage"
)

type App interface {
	Create(name, path string, migrationType string) error
	Up(path string) error
	Down(path string) error
	Redo(path string) error
	Status() error
	DBVersion() error
}

type Application struct {
	logger     logger.Logger
	SQLStorage storage.SQLStorage
}

var (
	ErrInvalidMigrationName = errors.New("invalid migration name")
	ErrMigrationFileExists  = errors.New("migration file already exists")
	ErrMigrationFailed      = errors.New("migration failed")
	ErrNoMigrationsFound    = errors.New("no migrations found")

	regGetVersion         = regexp.MustCompile(`^\d+`)
	regGetUpMigration     = regexp.MustCompile(`^.+_up\.sql$`)
	regGetDownMigration   = regexp.MustCompile(`^.+_down\.sql$`)
	regGetUpGoMigration   = regexp.MustCompile(`^.+_up\.go$`)
	regGetDownGoMigration = regexp.MustCompile(`^.+_down\.go$`)
)

func New(logger logger.Logger, SQLStorage storage.SQLStorage) *Application {
	return &Application{
		logger:     logger,
		SQLStorage: SQLStorage,
	}
}

func (app *Application) Create(name, filePath, migrationType string) error {
	if strings.TrimSpace(name) == "" {
		return ErrInvalidMigrationName
	}

	files, err := os.ReadDir(filePath)
	if err != nil {
		app.logger.Error("Failed to read directory: %v", err)
		return err
	}

	lastVersion := getLastVersion(files, app.logger)
	newVersion := lastVersion + 1

	if err := createMigrationFiles(filePath, newVersion, name, app.logger, migrationType); err != nil {
		app.logger.Error("Failed to create migration files: %v", err)
		return err
	}

	app.logger.Info("Created migration %s version %d", name, newVersion)
	return nil
}

func (app *Application) Up(filePath string) error {
	migrator := processes.New(app.SQLStorage, app.logger)
	migrations, err := getMigrations(filePath)
	if err != nil {
		app.logger.Error("Failed to get migrations: %v", err)
		return err
	}

	if len(migrations) == 0 {
		return ErrNoMigrationsFound
	}

	for _, migration := range migrations {
		migrator.Create(migration.Name, migration.Up, migration.Down, migration.UpGo, migration.DownGo)
	}

	ctx := context.Background()
	if err := migrator.Connect(ctx); err != nil {
		app.logger.Error("Failed to connect to database: %v", err)
		return err
	}
	defer migrator.Close(ctx)

	if err := migrator.Up(ctx); err != nil {
		app.logger.Error("Migration up failed: %v", err)
		return ErrMigrationFailed
	}

	return nil
}

func (app *Application) Down(filePath string) error {
	migrator := processes.New(app.SQLStorage, app.logger)
	migrations, err := getMigrations(filePath)
	if err != nil {
		app.logger.Error("Failed to get migrations: %v", err)
		return err
	}

	if len(migrations) == 0 {
		return ErrNoMigrationsFound
	}

	for _, migration := range migrations {
		migrator.Create(migration.Name, migration.Up, migration.Down, migration.UpGo, migration.DownGo)
	}

	ctx := context.Background()
	if err := migrator.Connect(ctx); err != nil {
		app.logger.Error("Failed to connect to database: %v", err)
		return err
	}
	defer migrator.Close(ctx)

	if err := migrator.Down(ctx); err != nil {
		app.logger.Error("Migration down failed: %v", err)
		return ErrMigrationFailed
	}

	return nil
}

func (app *Application) Redo(filePath string) error {
	migrator := processes.New(app.SQLStorage, app.logger)
	migrations, err := getMigrations(filePath)
	if err != nil {
		app.logger.Error("Failed to get migrations: %v", err)
		return err
	}

	if len(migrations) == 0 {
		return ErrNoMigrationsFound
	}

	for _, migration := range migrations {
		migrator.Create(migration.Name, migration.Up, migration.Down, migration.UpGo, migration.DownGo)
	}

	ctx := context.Background()
	if err := migrator.Connect(ctx); err != nil {
		app.logger.Error("Failed to connect to database: %v", err)
		return err
	}
	defer migrator.Close(ctx)

	if err := migrator.Redo(ctx); err != nil {
		app.logger.Error("Migration redo failed: %v", err)
		return ErrMigrationFailed
	}

	return nil
}

func (app *Application) Status() error {
	migrator := processes.New(app.SQLStorage, app.logger)
	ctx := context.Background()
	if err := migrator.Connect(ctx); err != nil {
		app.logger.Error("Failed to connect to database: %v", err)
		return err
	}
	defer migrator.Close(ctx)

	if err := migrator.Status(ctx); err != nil {
		app.logger.Error("Failed to get migration status: %v", err)
		return err
	}

	return nil
}

func (app *Application) DBVersion() error {
	migrator := processes.New(app.SQLStorage, app.logger)
	ctx := context.Background()
	if err := migrator.Connect(ctx); err != nil {
		app.logger.Error("Failed to connect to database: %v", err)
		return err
	}
	defer migrator.Close(ctx)

	if err := migrator.DBVersion(ctx); err != nil {
		app.logger.Error("Failed to get database version: %v", err)
		return err
	}

	return nil
}

func getLastVersion(files []os.DirEntry, logger logger.Logger) int {
	lastVersion := 0

	for _, file := range files {
		strVersion := regGetVersion.FindString(file.Name())

		if strVersion != "" {
			version, err := strconv.Atoi(strVersion)
			if err != nil {
				logger.Error("Failed to parse version: ", err)
				return -1
			}

			if version > lastVersion {
				lastVersion = version
			}
		}
	}

	return lastVersion
}

func createMigrationFiles(filePath string, version int, name string, logger logger.Logger, migrationType string) error {
	switch migrationType {
	case "sql":
		upFile := path.Join(filePath, fmt.Sprintf("%05d_%s_up.sql", version, name))
		err := os.WriteFile(upFile, []byte(""), 0o600)
		if err != nil {
			return err
		}
		logger.Info(upFile + " created_upFile")

		downFile := path.Join(filePath, fmt.Sprintf("%05d_%s_down.sql", version, name))
		err = os.WriteFile(downFile, []byte(""), 0o600)
		if err != nil {
			return err
		}
		logger.Info(downFile + " created_downFile")
	case "go":
		upFile := path.Join(filePath, fmt.Sprintf("%05d_%s_up.go", version, name))
		upContent := `package main

import (
	"context"
	"github.com/Edestus789/sql-migrator/storage"
)

func Up(ctx context.Context) error {
	db, ok := ctx.Value("db").(*storage.SQLStorage)
	if !ok {
		return fmt.Errorf("could not get database connection from context")
	}

	sql := "
		CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		username VARCHAR(255) NOT NULL,
		email VARCHAR(255) NOT NULL UNIQUE,
		created_at TIMESTAMP NOT NULL DEFAULT NOW()
		);"

	if err := db.Migrate(ctx, sql); err != nil {
		return fmt.Errorf("could not execute migration: %v", err)
	}

	fmt.Println("Migration Up applied: users table created")
	return nil
}
`
		err := os.WriteFile(upFile, []byte(upContent), 0o600)
		if err != nil {
			return err
		}
		logger.Info(upFile + " created_upFile")

		downFile := path.Join(filePath, fmt.Sprintf("%05d_%s_down.go", version, name))
		downContent := `package main

import (
	"context"
	"github.com/Edestus789/sql-migrator/storage"
)

func Down(ctx context.Context) error {
	db, ok := ctx.Value("db").(*storage.SQLStorage)
	if !ok {
		return fmt.Errorf("could not get database connection from context")
	}

	sql := "DROP TABLE IF EXISTS users;""

	if err := db.Migrate(ctx, sql); err != nil {
		return fmt.Errorf("could not execute migration: %v", err)
	}

	fmt.Println("Migration Down applied: users table dropped")
	return nil
}
`
		err = os.WriteFile(downFile, []byte(downContent), 0o600)
		if err != nil {
			return err
		}
		logger.Info(downFile + " created_downFile")
	default:
		return errors.New("unsupported migration type")
	}
	return nil
}

func getMigrations(filePath string) (map[int]*storage.Migration, error) {
	files, err := os.ReadDir(filePath)
	if err != nil {
		return nil, err
	}

	migrations := make(map[int]*storage.Migration)

	for _, file := range files {
		version, migrationName, err := parseFileName(file.Name())
		if err != nil {
			return nil, err
		}

		migration, err := processMigrationFile(filePath, file, version, migrationName)
		if err != nil {
			return nil, err
		}

		if migration != nil {
			if existingMigration, ok := migrations[version]; ok {
				mergeMigrations(existingMigration, migration)
			} else {
				migrations[version] = migration
			}
		}
	}

	return migrations, nil
}

func parseFileName(fileName string) (int, string, error) {
	strVersion := regGetVersion.FindString(fileName)
	if strVersion == "" {
		return 0, "", ErrInvalidMigrationName
	}

	version, err := strconv.Atoi(strVersion)
	if err != nil {
		return 0, "", err
	}

	parts := strings.Split(fileName, "_")
	if len(parts) < 3 {
		return 0, "", ErrInvalidMigrationName
	}

	migrationName := strings.Join(parts[1:len(parts)-1], "_")
	return version, migrationName, nil
}

func processMigrationFile(filePath string, file os.DirEntry, version int, migrationName string) (*storage.Migration, error) {
	filePathFull := path.Join(filePath, file.Name())

	switch {
	case regGetUpMigration.MatchString(file.Name()):
		sql, err := os.ReadFile(filePathFull)
		if err != nil {
			return nil, err
		}
		return &storage.Migration{
			Version: version,
			Name:    migrationName,
			Up:      string(sql),
		}, nil

	case regGetDownMigration.MatchString(file.Name()):
		sql, err := os.ReadFile(filePathFull)
		if err != nil {
			return nil, err
		}
		return &storage.Migration{
			Version: version,
			Name:    migrationName,
			Down:    string(sql),
		}, nil

	case regGetUpGoMigration.MatchString(file.Name()):
		return &storage.Migration{
			Version: version,
			Name:    migrationName,
			UpGo: func(ctx context.Context) error {
				return runGoMigration(filePath, file.Name())
			},
		}, nil

	case regGetDownGoMigration.MatchString(file.Name()):
		return &storage.Migration{
			Version: version,
			Name:    migrationName,
			DownGo: func(ctx context.Context) error {
				return runGoMigration(filePath, file.Name())
			},
		}, nil

	default:
		return nil, ErrInvalidMigrationName
	}
}

func mergeMigrations(existing, newMigration *storage.Migration) {
	if newMigration.Up != "" {
		existing.Up = newMigration.Up
	}
	if newMigration.Down != "" {
		existing.Down = newMigration.Down
	}
	if newMigration.UpGo != nil {
		existing.UpGo = newMigration.UpGo
	}
	if newMigration.DownGo != nil {
		existing.DownGo = newMigration.DownGo
	}
}

func runGoMigration(filePath, fileName string) error {
	// Проверяем, что fileName содержит только допустимые символы
	if !isSafeFilename(fileName) {
		return errors.New("invalid filename")
	}

	fullPath := path.Join(filePath, fileName)
	cmd := exec.Command("go", "run", fullPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func isSafeFilename(name string) bool {
	// Проверяем, что имя файла содержит только буквы, цифры, точки, подчеркивания и дефисы
	return regexp.MustCompile(`^[a-zA-Z0-9_.-]+$`).MatchString(name)
}
