package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/pflag"

	"gophkeeper/internal/config"
	httpx "gophkeeper/internal/http"
	"gophkeeper/internal/http/handlers"
	"gophkeeper/internal/logger"
	"gophkeeper/internal/repository/postgres"
	"gophkeeper/internal/service"

	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

func run() error {
	// --- flags ---
	fs := pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)
	fs.AddGoFlagSet(flag.CommandLine)

	fs.String("addr", "127.0.0.1:8080", "server address")
	fs.String("dsn", "", "PostgreSQL DSN")
	fs.String("jwt", "", "JWT secret")
	fs.String("log_level", "INFO", "log level (DEBUG, INFO, WARN, ERROR)")
	fs.String("read_timeout", "5s", "read timeout")
	fs.String("write_timeout", "5s", "write timeout")

	if err := fs.Parse(os.Args[1:]); err != nil {
		return err
	}

	// --- config ---
	cfg, err := config.Load(fs)
	if err != nil {
		return err
	}

	// --- logger ---
	if err := logger.Initialize(fs.Lookup("log_level").Value.String()); err != nil {
		return err
	}
	defer logger.Log.Sync()

	logger.Log.Info("logger initialized")

	// --- database ---
	db, err := sql.Open("pgx", cfg.DSN)
	if err != nil {
		return err
	}
	defer db.Close()

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	if err := db.Ping(); err != nil {
		return err
	}

	logger.Log.Info("connected to PostgreSQL")

	// --- repository ---
	storage := postgres.New(db)

	// --- services ---
	authSvc := service.NewAuthService(storage, cfg.JWTSecret)
	secSvc := service.NewSecretsService(storage)

	// --- handlers ---
	authH := handlers.NewAuthHandler(authSvc)
	secH := handlers.NewSecretsHandler(secSvc)

	// --- router ---
	router := httpx.NewRouter(
		cfg.JWTSecret,
		authH,
		secH,
	)

	// --- http server ---
	rt, _ := time.ParseDuration(cfg.ReadTimeout)
	wt, _ := time.ParseDuration(cfg.WriteTimeout)

	srv := &http.Server{
		Addr:         cfg.Addr,
		Handler:      router,
		ReadTimeout:  rt,
		WriteTimeout: wt,
	}

	errCh := make(chan error, 1)

	go func() {
		logger.Log.Info("server started", zap.String("addr", cfg.Addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	// --- graceful shutdown ---
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-stop:
		logger.Log.Info("shutdown signal received", zap.String("signal", sig.String()))

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			logger.Log.Error("server shutdown failed", zap.Error(err))
			return err
		}

		logger.Log.Info("server stopped gracefully")
		return nil

	case err := <-errCh:
		return err
	}
}
