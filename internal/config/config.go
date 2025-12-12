package config

import (
	"os"
)

type Config struct {
	DBURL string
}

func Load() *Config {
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		dbURL = "postgres://user:password@localhost:5432/dbname?sslmode=disable"
	}
	return &Config{
		DBURL: dbURL,
	}
}
