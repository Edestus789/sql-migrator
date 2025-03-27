package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	MigratorOpt *Migrator
	LoggerOpt   *Logger
}

type Migrator struct {
	DSN       string `mapstructure:"dsn"`
	Dir       string `mapstructure:"dir"`
	Type      string `mapstructure:"type"`
	TableName string `mapstructure:"table_name"`
}

type Logger struct {
	Level string `mapstructure:"level"`
}

func LoadConfig(configPath string) (*Config, error) {
	v := viper.New()

	// Если путь к конфигу явно не указан, ищем в текущей директории
	if configPath == "" {
		v.SetConfigName("config") // Имя файла без расширения
		v.AddConfigPath(".")      // Ищем в текущей директории
		v.SetConfigType("yaml")   // Поддерживаемые типы: yaml, json, toml и т.д.
	} else {
		v.SetConfigFile(configPath)
	}

	// Читаем конфиг
	if err := v.ReadInConfig(); err != nil {
		if configPath == "" {
			// Если конфиг не указан и не найден - возвращаем дефолтные значения
			return &Config{
				MigratorOpt: &Migrator{
					Dir:       "./migrations",
					Type:      "sql",
					TableName: "schema_migrations",
				},
				LoggerOpt: &Logger{
					Level: "info",
				},
			}, nil
		}
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	// Заменяем переменные окружения в значениях конфига
	for _, key := range v.AllKeys() {
		value := v.GetString(key)
		expanded := os.ExpandEnv(value)
		v.Set(key, expanded)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	return &cfg, nil
}
