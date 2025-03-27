package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Edestus789/sql-migrator/app"
	"github.com/Edestus789/sql-migrator/config"
	"github.com/Edestus789/sql-migrator/logger"
	"github.com/Edestus789/sql-migrator/storage"
)

var (
	configPath    string
	path          string
	database      string
	migrationType string
)

func main() {
	// Парсим флаги глобально
	flag.StringVar(&configPath, "config", "", "Path to config file")
	flag.StringVar(&path, "path", "", "Path to migrations directory")
	flag.StringVar(&database, "dsn", "", "Database connection string")
	flag.StringVar(&migrationType, "type", "", "Migration type: sql or go")
	flag.Parse()

	// Первый аргумент без флага - команда
	args := flag.Args()
	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}

	command := args[0]
	var migrationName string
	if command == "create" {
		if len(args) < 2 {
			fmt.Println("Error: migration name required for create command")
			printUsage()
			os.Exit(1)
		}
		migrationName = args[1]
	}

	// Загружаем конфиг
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Применяем приоритеты: флаги > конфиг > дефолты
	if path == "" {
		path = cfg.MigratorOpt.Dir
	}
	if database == "" {
		database = cfg.MigratorOpt.DSN
	}
	if migrationType == "" {
		migrationType = cfg.MigratorOpt.Type
	}

	// Валидация обязательных параметров
	if path == "" {
		fmt.Println("Error: migrations directory path not specified")
		printUsage()
		os.Exit(1)
	}

	if (command == "up" || command == "down" || command == "redo") && database == "" {
		fmt.Println("Error: database connection string not specified")
		printUsage()
		os.Exit(1)
	}

	// Инициализация приложения
	l := logger.New()
	db := storage.NewPostgresStorage(database, l)
	application := app.New(l, db)

	// Выполнение команды
	switch command {
	case "create":
		application.Create(migrationName, path, migrationType)
	case "up":
		application.Up(path)
	case "down":
		application.Down(path)
	case "redo":
		application.Redo(path)
	case "status":
		application.Status()
	case "dbversion":
		application.DBVersion()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`
Usage:
  gomigrator [flags] <command> [arguments]

Commands:
  create <name>    Create new migration
  up               Apply all migrations
  down             Rollback last migration
  redo             Redo last migration
  status           Show migration status
  dbversion        Show current database version

Flags:
  --config string  Path to config file (default "./config.yaml")
  --path string    Path to migrations directory
  --dsn string     Database connection string
  --type string    Migration type: sql or go
`)
}
