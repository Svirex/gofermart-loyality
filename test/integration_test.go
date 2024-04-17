//go:build integration
// +build integration

package test_test

import (
	"log"
	"testing"

	"github.com/Svirex/gofermart-loyality/test/testdb"
	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/postgres"
	"github.com/jackc/pgx/v5/stdlib"
)

func TestMain(m *testing.M) {
	testdb.Connect()
	defer testdb.Close()

	driver, err := postgres.WithInstance(stdlib.OpenDBFromPool(testdb.GetPool()), &postgres.Config{})
	if err != nil {
		log.Fatalf("create instance db for migrate: %v", err)
	}
	migration, err := migrate.NewWithDatabaseInstance(
		"file://../migrations", "postgres", driver)
	if err != nil {
		log.Fatalf("create migrate: %v", err)
	}
	err = migration.Up()
	if err != nil {
		log.Fatalf("migration up: %v", err)
	}
	m.Run()
	migration.Down()
}
