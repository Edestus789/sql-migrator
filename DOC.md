# SQL Migrator

## Описание
Инструмент для управления миграциями базы данных с поддержкой SQL и Go-миграций. Аналог решений из раздела "Database schema migration" awesome-go.

## Установка
```bash
go get github.com/Edestus789/sql-migrator/cmd/gomigrator
```

## Основные команды

### Создание миграции
```bash
gomigrator create <имя_миграции>
```
Создает пару файлов:
- `<версия>_<имя>_up.sql` (или `.go`)
- `<версия>_<имя>_down.sql` (или `.go`)

### Применение миграций
```bash
gomigrator up
```
Применяет все непримененные миграции в порядке версий.

### Откат миграции
```bash
gomigrator down
```
Откатывает последнюю примененную миграцию.

### Повтор миграции
```bash
gomigrator redo
```
Выполняет `down` + `up` для последней миграции.

### Просмотр статуса
```bash
gomigrator status
```
Выводит таблицу с информацией о миграциях:
```
| Название          | Статус  | Время               |
|-------------------|---------|---------------------|
| create_users      | success | 2023-01-01 12:00:00 |
| add_index         | pending | -                   |
```

### Версия БД
```bash
gomigrator dbversion
```
Показывает номер последней примененной миграции.

## Конфигурация
Настройки в `config.toml`:
```toml
[migrator]
dsn = "postgresql://user:pass@localhost:5432/db" # DSN подключения
dir = "./migrations"                            # Путь к миграциям
type = "sql"                                    # Тип миграций (sql/go)
table_name = "schema_migrations"                # Таблица для учета миграций

[logger]
level = "INFO"                                  # Уровень логирования
```

Или через аргументы:
```bash
gomigrator --dsn "postgresql://user:pass@localhost:5432/db" --path ./migrations up
```

## Форматы миграций

### SQL
Файлы:
- `0001_create_users_up.sql`
- `0001_create_users_down.sql`

Пример:
```sql
-- 0001_create_users_up.sql
CREATE TABLE users (
  id SERIAL PRIMARY KEY,
  name VARCHAR(255)
);

-- 0001_create_users_down.sql
DROP TABLE users;
```

### Go
Файлы:
- `0001_create_users_up.go`
- `0001_create_users_down.go`

Пример:
```go
// 0001_create_users_up.go
func Up(ctx context.Context) error {
    db := ctx.Value("db").(*storage.SQLStorage)
    return db.Migrate(ctx, `
        CREATE TABLE users (
            id SERIAL PRIMARY KEY,
            name VARCHAR(255)
        )
    `)
}

// 0001_create_users_down.go
func Down(ctx context.Context) error {
    db := ctx.Value("db").(*storage.SQLStorage)
    return db.Migrate(ctx, "DROP TABLE users")
}
```

## Особенности
- **Блокировки**: Использует `pg_advisory_lock` для безопасного параллельного выполнения
- **Транзакции**: Каждая миграция выполняется в отдельной транзакции
- **Логирование**: Подробные логи через zerolog
- **Тестирование**: Интеграционные тесты с Docker + PostgreSQL

## Пример workflow
1. Создать миграцию:
   ```bash
   gomigrator create add_users_table
   ```
2. Редактировать `migrations/*_add_users_table_*.sql`
3. Применить:
   ```bash
   gomigrator up
   ```
4. Проверить статус:
   ```bash
   gomigrator status
   ```

## Требования
- Go 1.22+
- PostgreSQL 9.6+


Да, **SQL Migrator** можно использовать как библиотеку в вашем Go-проекте. Вот как это сделать:

---

## **📦 Использование как библиотеки**

### **1. Импорт пакета**
Добавьте в ваш `go.mod`:
```bash
go get github.com/Edestus789/sql-migrator
```

### **2. Пример использования**
```go
package main

import (
	"context"
	"github.com/Edestus789/sql-migrator/app"
	"github.com/Edestus789/sql-migrator/logger"
	"github.com/Edestus789/sql-migrator/storage"
)

func main() {
	// 1. Инициализация логгера
	log := logger.New()

	// 2. Подключение к PostgreSQL
	dsn := "postgresql://user:pass@localhost:5432/db"
	pgStorage := storage.NewPostgresStorage(dsn, log)

	// 3. Создание экземпляра приложения
	migratorApp := app.New(log, pgStorage)

	// 4. Выполнение операций
	ctx := context.Background()

	// Пример: создание миграции
	migratorApp.Create("add_users", "./migrations", "sql")

	// Пример: применение миграций
	migratorApp.Up("./migrations")

	// Пример: проверка статуса
	migratorApp.Status()
}
```

---

## **🔧 Доступные методы API**
Основной интерфейс (`app.App`):
```go
type App interface {
    Create(name, path, migrationType string)  // Создать миграцию
    Up(path string)                          // Применить миграции
    Down(path string)                        // Откатить миграцию
    Redo(path string)                        // Перезапустить последнюю миграцию
    Status()                                 // Показать статус
    DBVersion()                              // Получить версию БД
}
```

---