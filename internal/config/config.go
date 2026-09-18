package config

import "os"

type Config struct {
	AppPort     string
	PostgresDSN string
	RedisAddr   string
	JWTSecret   string
}

func LoadConfig() Config {
	return Config{
		AppPort: getEnv("APP_PORT", "8080"),
		PostgresDSN: getEnv(
			"POSTGRES_DSN",
			"postgres://postgres:postgres@localhost:5432/urlshortener?sslmode=disable",
		),
		RedisAddr: getEnv("REDIS_ADDR", "localhost:6379"),
		JWTSecret: getEnv("JWT_SECRET", "change-me"),
	}
}

func getEnv(Key, fallback string) string {

	value := os.Getenv(Key)

	if value == "" {
		return fallback
	}

	return value
}
