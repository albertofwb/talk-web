package config

import (
	"fmt"
	"os"
)

type Config struct {
	DBHost      string
	DBPort      string
	DBUser      string
	DBPassword  string
	DBName      string
	RedisAddr   string
	JWTSecret   string
	TalkServerURL string
	Port        string
}

func Load() *Config {
	return &Config{
		DBHost:        getEnv("DB_HOST", "localhost"),
		DBPort:        getEnv("DB_PORT", "5432"),
		DBUser:        getEnv("DB_USER", "talk"),
		DBPassword:    getEnv("DB_PASSWORD", "talk"),
		DBName:        getEnv("DB_NAME", "talk"),
		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		JWTSecret:     getEnv("JWT_SECRET", "your-secret-key"),
		TalkServerURL: getEnv("TALK_SERVER_URL", "http://localhost:5000"),
		Port:          getEnv("PORT", "8080"),
	}
}

func (c *Config) DSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
