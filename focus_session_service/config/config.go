package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config stores all configuration for application
type Config struct {
	Environment string
	Server      ServerConfig
	MongoDB     MongoDBConfig
	JWT         JWTConfig
}

// ServerConfig stores configuration for web server
type ServerConfig struct {
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// MongoDBConfig stores configuration for MongoDB
type MongoDBConfig struct {
	URI      string
	Database string
	Timeout  time.Duration
}

// JWTConfig stores configuration for JWTConfig
type JWTConfig struct {
	SecretKey            string
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
}

// LoadConfigs loads configuration from environment variables
func LoadConfig() (*Config, error) {
	config := &Config{
		Environment: getEnv("ENVIRONMENT", "development"),
		Server: ServerConfig{
			Port:         getEnvAsInt("SERVER_PORT", 8080),
			ReadTimeout:  getEnvAsDuration("SERVER_READ_TIMEOUT", 10*time.Second),
			WriteTimeout: getEnvAsDuration("SERVER_WRITE_TIMEOUT", 10*time.Second),
		},
		MongoDB: MongoDBConfig{
			URI:      getEnv("MONGODB_URI", "mongodb://localhost:27017"),
			Database: getEnv("MONGODB_DATABASE", "user_service"),
			Timeout:  getEnvAsDuration("MONGODB_TIMEOUT", 10*time.Second),
		},
		JWT: JWTConfig{
			SecretKey:            getEnv("JWT_SECRET_KEY", "your-secret-key-must-be-at-least-32-characters"),
			AccessTokenDuration:  getEnvAsDuration("JWT_ACCESS_TOKEN_DURATION", 24*time.Hour),
			RefreshTokenDuration: getEnvAsDuration("JWT_REFRESH_TOKEN_DURATION", 7*24*time.Hour),
		},
	}

	// Validate JWT secret key
	if len(config.JWT.SecretKey) < 32 {
		return nil, fmt.Errorf("JWT_SECRET_KEY must be at least 32 characters long")
	}

	return config, nil
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	valueStr := getEnv(key, "")
	if value, err := time.ParseDuration(valueStr); err == nil {
		return value
	}

	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}
