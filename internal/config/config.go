package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	ServerConfig
	DbConfig
}

type ServerConfig struct {
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
			Address:     getEnvString("SERVER_ADDRESS", "0.0.0.0:3353"),
			MaxByteSend: getEnvInt("MAX_BYTE_SEND", 1000),
			TokenTTL:    getEnvTime("TokenTTL", time.Hour),
			StoragePath: getEnvString("STORAGE_PATH", "./serverStorage/"),
			LogFilePath: getEnvString("LOGFILE", "./internal/logger/"),
		},
		DbConfig: DbConfig{
			Host:     getEnvString("POSTGRES_HOST", "localhost"),
			Port:     getEnvString("ADDRESS_FOR_DB", "3354"),
			Username: getEnvString("POSTGRES_USERNAME", "kulakov"),
			Password: getEnvString("POSTGRES_PASSWORD", "1234"),
			Database: getEnvString("POSTGRES_NAME", "fileStorage"),
		},
	}
}

func getEnvString(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if value, exists := os.LookupEnv(key); exists {
		if value, err := strconv.Atoi(value); err == nil {
			return value
		}

	}
	return defaultVal
}

func getEnvTime(key string, defaultVal time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists {
		if value, err := time.ParseDuration(value); err == nil {
			return value
		}
	}
	return defaultVal
}
