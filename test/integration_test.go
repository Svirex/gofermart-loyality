//go:build integration
// +build integration

package test

import (
	"os"
	"testing"

	"github.com/Svirex/gofermart-loyality/test/testdb"
	_ "github.com/golang-migrate/migrate/source/file"
	"go.uber.org/zap"
)

var logger *zap.SugaredLogger

func TestMain(m *testing.M) {
	testdb.Init()
	defer testdb.Close()

	testdb.MigrateUp()
	code := m.Run()
	testdb.MigrateDown()
	os.Exit(code)
}
