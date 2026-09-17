package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds every environment-driven setting the server needs.
// Loading it once at startup means the rest of the app never touches
// os.Getenv directly, which keeps configuration in one place.
type Config struct {
	Port           string
	MongoURI       string
	MongoDBName    string
	RedisAddr      string
	RedisPassword  string
	RedisDB        int
	JWTSecret      string
	AllowedOrigin  string
	TokenTTLHours  int
}

func Load() Config {
	// .env is optional: in production (Render/Railway/etc.) env vars are
	// injected by the platform, so a missing .env file is not an error.
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, relying on real environment variables")
	}

	return Config{
		Port:          getEnv("PORT", "8080"),
		MongoURI:      getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDBName:   getEnv("MONGO_DB_NAME", "polling_app"),
		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       0,
		JWTSecret:     getEnv("JWT_SECRET", "dev-secret-change-me"),
		AllowedOrigin: getEnv("ALLOWED_ORIGIN", "http://localhost:5173"),
		TokenTTLHours: 72,
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
