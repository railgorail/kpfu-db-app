package config

import (
	"os"
)

// Config holds the application configuration.
type Config struct {
	DBURL string
}

// Load returns a new Config struct.
func Load() *Config {
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		dbURL = "postgres://user:password@localhost:5432/dbname?sslmode=disable"
	}
	return &Config{
		DBURL: dbURL,
	}
}
