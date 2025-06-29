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
	PusherAppID            string
	PusherKey              string
	PusherSecret           string
	PusherCluster          string
	GoogleClientID         string
	GoogleClientSecret     string
	GithubId               string
	GithubSecret           string
	SessionSecret          string
	BaseUrl                string
	ClientBaseUrl          string
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
		PusherAppID:            getEnv("PUSHER_APP_ID", ""),
		PusherKey:              getEnv("PUSHER_KEY", ""),
		PusherSecret:           getEnv("PUSHER_SECRET", ""),
		PusherCluster:          getEnv("PUSHER_CLUSTER", ""),
		GoogleClientID:         getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret:     getEnv("GOOGLE_CLIENT_SECRET", ""),
		GithubId:               getEnv("GITHUB_ID", ""),
		GithubSecret:           getEnv("GITHUB_SECRET", ""),
		SessionSecret:          getEnv("SESSION_SECRET", ""),
		BaseUrl:                getEnv("BASE_URL", ""),
		ClientBaseUrl:          getEnv("CLIENT_BASE_URL", ""),
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
