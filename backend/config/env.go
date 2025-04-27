package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	PublicHost             string
	Port                   string
	JWTExpirationInSeconds int64
	JWTSecret              string
	OpenAIKey              string
	DatabaseUrl            string
	RemoteDBUrl            string
	FrontendUrl            string
	IsProduction           bool
}

var Envs = initConfig()

func initConfig() Config {
	godotenv.Load()

	return Config{
		PublicHost:             getEnv("PUBLIC_HOST", "http://localhost"),
		Port:                   getEnv("PORT", "8080"),
		JWTExpirationInSeconds: getEnvAsInt("JWT_EXPIRATION_IN_SECONDS", 3600*24*7),
		JWTSecret:              getEnv("JWT_SECRET", "not-so-secret"),
		OpenAIKey:              getEnv("OPENAI_API_KEY", "not-so-secret"),
		DatabaseUrl:            getEnv("DATABASE_URL", "mysql"),
		RemoteDBUrl:            getEnv("REMOTE_DB_URL", ""),
		FrontendUrl:            getEnv("FRONTEND_URL", "http://localhost:3000"),
		IsProduction:           getEnvAsBool("IS_PRODUCTION", false),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return fallback
}

func getEnvAsInt(key string, fallback int64) int64 {
	if value, ok := os.LookupEnv(key); ok {
		i, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fallback
		}
		return i
	}

	return fallback

}

func getEnvAsBool(key string, fallback bool) bool {
	if value, ok := os.LookupEnv(key); ok {
		b, err := strconv.ParseBool(value)
		if err != nil {
			return fallback
		}
		return b
	}

	return fallback
}
