package testdb

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

var db *pgxpool.Pool

func GetPool() *pgxpool.Pool {
	if db == nil {
		log.Fatalf("db not init")
	}
	return db
}

func Connect() {
	var db_url string
	var exists bool
	if db_url, exists = os.LookupEnv("DB_URL"); !exists {
		log.Fatalf("connect string DB_URL not exists")
	}
	var err error
	db, err = pgxpool.New(context.Background(), db_url)
	if err != nil {
		log.Fatalf("coulnd't connect to db: %s", db_url)
	}
}

func Close() {
	db.Close()
}

func Truncate() {

}
