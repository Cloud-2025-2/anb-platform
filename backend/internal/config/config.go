package config

import (
	"log"
	"os"
	"strconv"
)

type Config struct {
	AppPort          string
	JWTSecret        string
	JWTExpireMinutes int

	// DB
	PostgresURL string
	// Redis
	RedisAddr     string
	RedisPassword string
}

func atoiEnv(k string, def int) int {
	if v := os.Getenv(k); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func Load() *Config {
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8000"
	}
	// DATABASE_URL opcional; si no está, construimos desde POSTGRES_*
	pgURL := os.Getenv("DATABASE_URL")
	if pgURL == "" {
		host := getenv("POSTGRES_HOST", "localhost")
		user := getenv("POSTGRES_USER", "proyecto1")
		pass := getenv("POSTGRES_PASSWORD", "proyecto1")
		db   := getenv("POSTGRES_DB", "db")
		portP := getenv("POSTGRES_PORT", "5432")
		pgURL = "host=" + host + " user=" + user + " password=" + pass +
			" dbname=" + db + " port=" + portP + " sslmode=disable"
	}

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Println("⚠️  JWT_SECRET no definido (usa uno seguro en prod)")
	}

	return &Config{
		AppPort:          port,
		JWTSecret:        secret,
		JWTExpireMinutes: atoiEnv("JWT_EXPIRE_MINUTES", 60),
		PostgresURL:      pgURL,
		RedisAddr:        getenv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:    os.Getenv("REDIS_PASSWORD"),
	}
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
