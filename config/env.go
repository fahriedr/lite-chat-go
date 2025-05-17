package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	MongoUrl               string
	Database               string
	Port                   string
	JWTExpirationInSeconds int64
	JWTSecret              string
	Robohash               string
}

var Envs = initConfig()

func initConfig() Config {
	godotenv.Load()
	return Config{
		MongoUrl:               getEnv("MONGO_URL", ""),
		Database:               getEnv("MONGO_DB_NAME", ""),
		Port:                   getEnv("PORT", ""),
		JWTSecret:              getEnv("JWT_SECRET", "not-secret-secret-anymore?"),
		JWTExpirationInSeconds: getEnvInt("JWT_EXP", 3600*24*7),
		Robohash:               getEnv("ROBOHASH_URL", ""),
	}

}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return fallback
}

func getEnvInt(key string, fallback int64) int64 {
	if value, ok := os.LookupEnv(key); ok {
		i, err := strconv.ParseInt(value, 10, 64)

		if err != nil {
			return fallback
		}

		return i
	}

	return fallback
}
