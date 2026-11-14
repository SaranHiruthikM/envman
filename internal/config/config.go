package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

const minJWTSecretLen = 32

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
	jwtSecret := getEnvRequired("JWT_SECRET")
	if len(jwtSecret) < minJWTSecretLen {
		panic(fmt.Sprintf("JWT_SECRET must be at least %d characters (got %d)", minJWTSecretLen, len(jwtSecret)))
	}
	return &Config{
		App: AppConfig{
			Port:      getEnvOptional("PORT", "8080"),
			JWTSecret: jwtSecret,
		},
		DB: DBConfig{DBUrl: getEnvRequired("DB_URL")},
	}
}

func getEnvRequired(key string) string {
	v, ok := os.LookupEnv(key)
	if !ok {
		panic("required environment variable not set: " + key)
	}
	return v
}

func getEnvOptional(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}
