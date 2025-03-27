package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Migrator struct {
		DSN       string `mapstructure:"dsn"`
		Dir       string `mapstructure:"dir"`
		Type      string `mapstructure:"type"`
		TableName string `mapstructure:"table_name"`
	} `mapstructure:"migrator"`
	Logger struct {
		Level string `mapstructure:"level"`
	} `mapstructure:"logger"`
}

func LoadConfig(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.SetConfigType(filepath.Ext(configPath)[1:]) // auto-detect format

	// Enable env vars expansion
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config: %w", err)
	}

	// Expand environment variables in config
	for _, key := range viper.AllKeys() {
		value := viper.GetString(key)
		expanded := os.ExpandEnv(value)
		viper.Set(key, expanded)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("config unmarshal error: %w", err)
	}

	return &cfg, nil
}
