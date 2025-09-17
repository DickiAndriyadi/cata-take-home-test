package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

type Config struct {
	MySQLDSN            string
	RedisAddr           string
	RedisPassword       string
	ServerPort          string
	APITimeout          time.Duration
	APIBaseURL          string
	CacheTTL            time.Duration
	SyncInterval        time.Duration
	MaxRetryAttempts    int
	RetryInitialBackoff time.Duration
}

func Load(logger *zap.Logger) Config {
	env := getEnv("GO_ENV", "development")

	// Load .env file if it exists
	if err := godotenv.Load(".env." + env); err != nil {
		if logger != nil {
			logger.Info("No .env file found. Using environment variables.")
		}
	}

	return Config{
		MySQLDSN:            getEnv("MYSQL_DSN", "root:password@tcp(localhost:3306)/pokemondb?parseTime=true"),
		RedisAddr:           getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:       getEnv("REDIS_PASSWORD", ""),
		ServerPort:          getEnv("SERVER_PORT", "8080"),
		APITimeout:          mustParseDuration(getEnv("API_TIMEOUT", "5s")),
		APIBaseURL:          getEnv("API_BASE_URL", "https://pokeapi.co/api/v2"),
		CacheTTL:            mustParseDuration(getEnv("CACHE_TTL", "5m")),
		SyncInterval:        mustParseDuration(getEnv("SYNC_INTERVAL", "15m")),
		MaxRetryAttempts:    mustParseInt(getEnv("MAX_RETRY_ATTEMPTS", "5")),
		RetryInitialBackoff: mustParseDuration(getEnv("RETRY_INITIAL_BACKOFF", "1s")),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func mustParseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		panic(err)
	}
	return d
}

func mustParseInt(s string) int {
	v, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}
	return v
}
