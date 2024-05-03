package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"
	"time"

	"github.com/Svirex/gofermart-loyality/internal/adapters/api"
	adapterspg "github.com/Svirex/gofermart-loyality/internal/adapters/postgres"
	"github.com/Svirex/gofermart-loyality/internal/common"
	"github.com/Svirex/gofermart-loyality/internal/config"
	"github.com/Svirex/gofermart-loyality/internal/core/services"
	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/postgres"
	_ "github.com/golang-migrate/migrate/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	cfg, err := config.Parse()
	if err != nil {
		log.Fatal(err)
	}

	config := zap.Config{
		Level:            zap.NewAtomicLevelAt(zapcore.DebugLevel),
		Development:      false,
		Encoding:         "json",
		EncoderConfig:    zap.NewProductionEncoderConfig(),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	l, err := config.Build()
	if err != nil {
		log.Fatalf("couldn't init zap logger")
	}
	logger := common.Logger(l.Sugar())

	dbpool, err := pgxpool.New(context.Background(), cfg.DatabaseURI)
	if err != nil {
		logger.Fatalf("create new pgxpool: %s, err: %v", cfg.DatabaseURI, err)
	}
	err = dbpool.Ping(context.Background())
	if err != nil {
		logger.Fatalf("db ping error: %v", err)
	}

	db := stdlib.OpenDBFromPool(dbpool)
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		logger.Fatalf("create instance db for migrate: %v", err)
	}
	currentDir, err := os.Getwd()
	if err != nil {
		logger.Fatalf("getwd: %v", err)
	}
	migrationPath := path.Join(strings.Replace(currentDir, "cmd/gophermart", "", 1), "migrations")
	migration, err := migrate.NewWithDatabaseInstance(
		"file://"+migrationPath, "postgres", driver)
	if err != nil {
		logger.Fatalf("create migrate: %v", "err", err)
	}

	err = migration.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		logger.Fatalf("migration up error ", "err=", err)
	}

	serverCtx, serverCancel := context.WithCancel(context.Background())

	authRepo := adapterspg.NewAuthRepository(dbpool)
	auth, err := services.NewAuthService(authRepo, 80, 8, 10, cfg.SecretKey)
	if err != nil {
		logger.Fatalf("auth service create: ", err)
	}

	ordersRepo := adapterspg.NewOrdersRepository(dbpool, logger)
	orders, err := services.NewOrderService(dbpool, ordersRepo, logger, 20, cfg.AccrualSystemAddress, 100*time.Millisecond, 20, 2*time.Second)
	if err != nil {
		logger.Fatalf("orders service create: ", err)
	}
	defer orders.Shutdown()

	balanceRepo := adapterspg.NewBalanceRepository(dbpool, logger)
	balance := services.NewBalanceService(balanceRepo)

	withdrawRepo := adapterspg.NewWithdrawRepository(dbpool, logger)
	withdraw := services.NewWithdrawService(withdrawRepo)

	api := api.NewAPI(auth, orders, balance, withdraw, logger)

	server := &http.Server{
		Addr:        cfg.RunAddress,
		Handler:     api.Routes(),
		BaseContext: func(net.Listener) context.Context { return serverCtx },
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		s := <-signalChan
		logger.Info("Received os.Signal. Try graceful shutdown.", "signal=", s)

		shutdownCtx, shutdownCancel := context.WithTimeout(serverCtx, 30*time.Second)
		defer shutdownCancel()

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				logger.Error("Gracelful shutdown timeout. Force shutdown")
				os.Exit(1)
			}
		}()

		err := server.Shutdown(shutdownCtx)
		if err != nil {
			logger.Error("Error while shutdown", "err", err)
			os.Exit(1)
		}

		serverCancel()

		logger.Info("Server shutdowned")
	}()
	logger.Info("Starting listen and serve...", "addr=", server.Addr)
	err = server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}

	<-serverCtx.Done()
}
