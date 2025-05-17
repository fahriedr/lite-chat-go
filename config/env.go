package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	MongoUrl string
	Database string
	Port     string
}

var Envs = initConfig()

func initConfig() Config {
	godotenv.Load()
	return Config{
		MongoUrl: getEnv("MONGO_URL", ""),
		Database: getEnv("MONGO_DB_NAME", ""),
		Port:     getEnv("PORT", ""),
	}

}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return fallback
}
