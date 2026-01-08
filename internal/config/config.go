package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	CORS     CORSConfig
}

type ServerConfig struct {
	Port     string
	LogLevel string
}

type DatabaseConfig struct {
	Path string
}

type CORSConfig struct {
	AllowedOrigins []string
}

func LoadConfig() *Config {
	// Load .env file if it exists
	_ = godotenv.Load()

	// Set default values
	config := &Config{
		Server: ServerConfig{
			Port:     getEnv("PORT", "3000"),
			LogLevel: getEnv("LOG_LEVEL", "debug"),
		},
		Database: DatabaseConfig{
			Path: getEnv("DB_PATH", "./data/greenlabelai.db"),
		},
		CORS: CORSConfig{
			AllowedOrigins: getEnvAsSlice("ALLOWED_ORIGINS", []string{"*"}, ","),
		},
	}

	// Ensure data directory exists
	if err := os.MkdirAll("./data", 0755); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}

	return config
}

// Helper function to read an environment variable or return a default value
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// Helper function to read an environment variable into a string slice
type getEnvAsSliceOptions struct {
	Separator string
}

func getEnvAsSlice(name string, defaultVal []string, separator string) []string {
	valStr := getEnv(name, "")
	if valStr == "" {
		return defaultVal
	}

	if len(valStr) == 0 {
		return []string{}
	}

	return splitString(valStr, separator)
}

// splitString splits a string by separator and returns a slice of strings
type splitStringOptions struct {
	TrimSpace bool
}

func splitString(str, separator string) []string {
	if len(str) == 0 {
		return []string{}
	}

	result := strings.Split(str, separator)
	for i := range result {
		result[i] = strings.TrimSpace(result[i])
	}

	return result
}

// Helper function to read an environment variable into an int
func getEnvAsInt(name string, defaultVal int) int {
	valueStr := getEnv(name, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultVal
}

// Helper function to read an environment variable into a bool
func getEnvAsBool(name string, defaultVal bool) bool {
	valStr := getEnv(name, "")
	if val, err := strconv.ParseBool(valStr); err == nil {
		return val
	}
	return defaultVal
}
