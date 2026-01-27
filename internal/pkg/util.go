package pkg

import (
	"log"

	"github.com/jackc/pgx/v4/pgxpool"
)

func NewDBConfig(databaseURL string) *pgxpool.Config {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		log.Fatalf("Failed to parse database URL: %v", err)
	}
	return config
}
