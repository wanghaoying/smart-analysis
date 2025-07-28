package config

import (
	"os"
)

type Config struct {
	Port        string
	DatabaseURL string
	JWTSecret   string
	UploadPath  string
	MaxFileSize int64
	OpenAIKey   string
	HunyuanKey  string
}

func Load() *Config {
	return &Config{
		Port:        getEnv("PORT", "8080"),
		DatabaseURL: getEnv("DATABASE_URL", "user:password@tcp(localhost:3306)/smart_analysis?charset=utf8mb4&parseTime=True&loc=Local"),
		JWTSecret:   getEnv("JWT_SECRET", "your-secret-key"),
		UploadPath:  getEnv("UPLOAD_PATH", "./uploads"),
		MaxFileSize: 500 * 1024 * 1024, // 500MB
		OpenAIKey:   getEnv("OPENAI_API_KEY", ""),
		HunyuanKey:  getEnv("HUNYUAN_API_KEY", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
