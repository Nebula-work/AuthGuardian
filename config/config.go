package config

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	// Server
	ServerPort string

	// MongoDB
	MongoURI     string
	DatabaseName string

	// JWT
	JWTSecret         string
	JWTExpirationTime time.Duration

	// OAuth2
	GoogleClientID     string
	GoogleClientSecret string
	GitHubClientID     string
	GitHubClientSecret string

	// CORS
	CORSAllowOrigins         string
	JWTRefreshExpirationTime time.Duration
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	// Load .env file if it exists
	godotenv.Load()

	config := &Config{
		ServerPort:               getEnv("PORT", "5000"),
		MongoURI:                 getEnv("MONGO_URI", "mongodb+srv://admin:admin@cluster0test.rwfvm.mongodb.net"),
		DatabaseName:             getEnv("DATABASE_NAME", "rbac_system"),
		JWTSecret:                getEnv("JWT_SECRET", "your-secret-key"),
		JWTExpirationTime:        time.Duration(getEnvAsInt("JWT_EXPIRATION_HOURS", 24)) * time.Hour,
		GoogleClientID:           getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret:       getEnv("GOOGLE_CLIENT_SECRET", ""),
		GitHubClientID:           getEnv("GITHUB_CLIENT_ID", ""),
		GitHubClientSecret:       getEnv("GITHUB_CLIENT_SECRET", ""),
		CORSAllowOrigins:         getEnv("CORS_ALLOW_ORIGINS", "http://localhost:3000"),
		JWTRefreshExpirationTime: time.Duration(getEnvAsInt("JWT_REFRESH_EXPIRATION_HOURS", 720)) * time.Hour, // 30 days
	}

	return config, nil
}

// Helper function to get environment variables
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// Helper function to get environment variables as integers
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}

	value := 0
	_, err := fmt.Sscanf(valueStr, "%d", &value)
	if err != nil {
		return defaultValue
	}

	return value
}
