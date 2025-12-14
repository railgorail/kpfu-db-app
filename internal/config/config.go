package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBURL string
}

func Load() *Config {
	_ = godotenv.Load() // try to load .env if present
	dbURL := os.Getenv("DB_URL")
	log.Println("DB_URL: ", dbURL)
	if dbURL == "" {
		dbURL = "postgres://user:password@localhost:5432/dbname?sslmode=disable"
	}
	return &Config{
		DBURL: dbURL,
	}
}
