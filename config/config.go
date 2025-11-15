package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	MONGODB_URL    string
	CloudinaryURL string
	JWTSecret     string
	AdminEmail    string
	AdminPass     string
}

func Load() *Config {
	_ = godotenv.Load()
	return &Config{
		MONGODB_URL:      getEnv("MONGODB_URL", ""),
		CloudinaryURL: getEnv("CLOUDINARY_URL", ""),
		JWTSecret:     getEnv("SECRET_KEY", "change-me"),
		AdminEmail:    getEnv("ADMIN_EMAIL", ""),
		AdminPass:     getEnv("ADMIN_PASSWORD", ""),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}