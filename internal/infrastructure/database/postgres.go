package database

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgres() *pgxpool.Pool {
	dbUrl := os.Getenv("DATABASE_URL")

	dbPool, err := pgxpool.New(context.Background(), dbUrl)
	if err != nil {
		log.Fatal("Unable to connected database:", err)
	}

	err = dbPool.Ping(context.Background())
	if err != nil {
		log.Fatal("Datavase not responding:", err)
	}
	log.Println("Connected to database.")

	return dbPool
}
