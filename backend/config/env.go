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
	GoogleClientID         string
	SMTPHost               string
	SMTPPort               string
	SMTPUsername           string
	SMTPPassword           string
	SMTPFromEmail          string
	SMTPFromName           string
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
		GoogleClientID:         getEnv("GOOGLE_CLIENT_ID", ""),
		SMTPHost:               getEnv("SMTP_HOST", ""),
		SMTPPort:               getEnv("SMTP_PORT", "587"),
		SMTPUsername:           getEnv("SMTP_USERNAME", ""),
		SMTPPassword:           getEnv("SMTP_PASSWORD", ""),
		SMTPFromEmail:          getEnv("SMTP_FROM_EMAIL", ""),
		SMTPFromName:           getEnv("SMTP_FROM_NAME", "Grão Certo"),
		DatabaseUrl:            getEnv("DATABASE_URL", "postgresql://postgres:postgres@localhost:5432/wallet_tracker"),
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
