package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	App AppConfig
	DB  DBConfig
}

type AppConfig struct {
	Port      string
	JWTSecret string
}

type DBConfig struct {
	DBUrl string
}

func Load() *Config {
	_ = godotenv.Load()
	return &Config{
		App: AppConfig{
			Port:      getEnvOptional("PORT", "8080"),
			JWTSecret: getEnvRequired("JWT_SECRET"),
		},
		DB: DBConfig{
			DBUrl: getEnvRequired("DB_URL"),
		},
	}
}

func getEnvRequired(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		panic("required environment variable not set: " + key)
	}

	return value
}

func getEnvOptional(key, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return fallback
	}

	return value
}
