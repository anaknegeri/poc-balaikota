package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the application
type Config struct {
	Server          ServerConfig
	Database        DatabaseConfig
	DataDirectories DirectoryConfig
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
	BaseURL         string `mapstructure:"base_url"`
}

// DatabaseConfig holds database-related configuration
type DatabaseConfig struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type DirectoryConfig struct {
	Root               string
	AlertDir           string
	PeopleCountDir     string
	FaceRecognitionDir string
	StreamDir          string
	VehicleCountDir    string
}

// Load loads configuration from environment variables
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port:            getEnv("PORT", "3002"),
			ReadTimeout:     getDurationEnv("SERVER_READ_TIMEOUT", 10*time.Second),
			WriteTimeout:    getDurationEnv("SERVER_WRITE_TIMEOUT", 30*time.Second),
			ShutdownTimeout: getDurationEnv("SERVER_SHUTDOWN_TIMEOUT", 10*time.Second),
		},
		Database: DatabaseConfig{
			DSN:             getEnv("DATABASE_URL", "host=localhost user=postgres password=password dbname=people_counting port=54320 sslmode=disable TimeZone=Asia/Jakarta"),
			MaxOpenConns:    getIntEnv("DATABASE_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getIntEnv("DATABASE_MAX_IDLE_CONNS", 10),
			ConnMaxLifetime: getDurationEnv("DATABASE_CONN_MAX_LIFETIME", 5*time.Minute),
		},
		DataDirectories: DirectoryConfig{
			Root:               "data",
			AlertDir:           "alert",
			PeopleCountDir:     "people-count",
			FaceRecognitionDir: "face_log",
			VehicleCountDir:    "car-count",
			StreamDir:          "stream",
		},
	}
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getIntEnv gets an environment variable as int or returns a default value
func getIntEnv(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}

	return intValue
}

// getDurationEnv gets an environment variable as duration or returns a default value
func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	duration, err := time.ParseDuration(value)
	if err != nil {
		return defaultValue
	}

	return duration
}
