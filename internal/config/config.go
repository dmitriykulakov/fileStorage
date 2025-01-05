package config

import (
	"os"
	"strconv"
	"time"

	"go.uber.org/zap/zapcore"
)

type Config struct {
	ServerConfig
	DbConfig
}

type ServerConfig struct {
	Level       zapcore.Level
	Address     string
	MaxByteSend int
	TokenTTL    time.Duration
	StoragePath string
	LogFilePath string
}

type DbConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	Database string
}

func NewConfig() *Config {
	return &Config{
		ServerConfig: ServerConfig{
			Level:       getEnv("SERVER_MODE", zapcore.InfoLevel).(zapcore.Level),
			Address:     getEnv("SERVER_ADDRESS", "0.0.0.0:3353").(string),
			MaxByteSend: getEnv("MAX_BYTE_SEND", 1000).(int),
			TokenTTL:    getEnv("TokenTTL", time.Hour).(time.Duration),
			StoragePath: getEnv("STORAGE_PATH", "./serverStorage/").(string),
			LogFilePath: getEnv("LOGFILE", "./internal/logger/").(string),
		},
		DbConfig: DbConfig{
			Host:     getEnv("POSTGRES_HOST", "localhost").(string),
			Port:     getEnv("ADDRESS_FOR_DB", "3354").(string),
			Username: getEnv("POSTGRES_USERNAME", "kulakov").(string),
			Password: getEnv("POSTGRES_PASSWORD", "1234").(string),
			Database: getEnv("POSTGRES_NAME", "fileStorage").(string),
		},
	}
}

func getEnv(key string, defaultVal interface{}) interface{} {
	if value, exists := os.LookupEnv(key); exists {
		switch defaultVal.(type) {
		case string:
			return value
		case zapcore.Level:
			if value == "true" {
				return true
			}
		case int:
			if value, err := strconv.Atoi(value); err == nil {
				return value
			}
		case time.Duration:
			if value, err := time.ParseDuration(value); err == nil {
				return value
			}
		}
	}
	return defaultVal
}
