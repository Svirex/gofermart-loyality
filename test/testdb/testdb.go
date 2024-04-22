package testdb

import (
	"context"
	"errors"
	"log"
	"os"

	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var dbpool *pgxpool.Pool
var logger *zap.SugaredLogger

var migration *migrate.Migrate

func GetPool() *pgxpool.Pool {
	if dbpool == nil {
		log.Fatalf("db not init")
	}
	return dbpool
}

func initLogger() {
	config := zap.Config{
		Level:            zap.NewAtomicLevelAt(zapcore.DebugLevel),
		Development:      true,
		Encoding:         "json",
		EncoderConfig:    zap.NewProductionEncoderConfig(),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	l, err := config.Build()
	if err != nil {
		log.Fatalf("couldn't init zap logger")
	}
	logger = l.Sugar()
}

func initMigration() {
	db := stdlib.OpenDBFromPool(dbpool)
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		logger.Fatalf("create instance db for migrate: %v", "err", err)
	}
	migration, err = migrate.NewWithDatabaseInstance(
		"file:///integration-test/migrations", "postgres", driver)
	if err != nil {
		logger.Fatalf("create migrate: %v", "err", err)
	}
}

func Init() {
	if logger == nil {
		initLogger()
	}
	if dbpool == nil {
		Connect()
	}
	if migration == nil {
		initMigration()
	}
}

func Connect() {
	var db_url string
	var exists bool
	if db_url, exists = os.LookupEnv("DB_URL"); !exists {
		log.Fatalf("connect string DB_URL not exists")
	}
	var err error
	dbpool, err = pgxpool.New(context.Background(), db_url)
	if err != nil {
		log.Fatalf("create new pgxpool: %s, err: %v", db_url, err)
	}
	err = dbpool.Ping(context.Background())
	if err != nil {
		log.Fatalf("db ping error: %v", err)
	}
	log.Println("DB Connected")
}

func Close() {
	dbpool.Close()
}

func MigrateUp() {
	err := migration.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		logger.Fatalf("migration up error ", "err=", err)
	}
}

func MigrateDown() {
	err := migration.Down()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		logger.Info("couldn't down migration ", "err=", err)
	}
}

func Truncate() error {
	_, err := dbpool.Exec(context.Background(), "TRUNCATE TABLE users RESTART IDENTITY;")
	if err != nil {
		logger.Error("couldn't truncate tables ", err)
		return err
	}
	return nil
}
