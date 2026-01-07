// Command server starts the GophKeeper HTTP API server.
// The server provides user authentication, secure secret storage,
// and synchronization functionality using PostgreSQL as storage
// and JWT for authorization.
package main

import (
	"context"
	"crypto/tls"
	"database/sql"
	"errors"
	"flag"
	"fmt"
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

	"gophkeeper/internal/crypto/envelope"

	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

// main is the entry point of the server application.
// It delegates all initialization logic to run and
// terminates the process on fatal startup errors.
func main() {
	if err := run(); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

// run initializes and starts the HTTP server.
// It performs the following steps:
//
//   - parses command-line flags
//   - loads application configuration
//   - initializes logging
//   - establishes a database connection
//   - configures repositories, services, and HTTP handlers
//   - starts the HTTP or HTTPS server
//   - handles graceful shutdown on system signals
//
// The function blocks until the server is stopped and
// returns an error if startup or runtime failure occurs.
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
	fs.String("master_key", "", "master key for data encryption (base64)")
	fs.String("tls_cert", "", "path to TLS certificate")
	fs.String("tls_key", "", "path to TLS key")

	if err := fs.Parse(os.Args[1:]); err != nil {
		return err
	}

	// --- config ---
	cfg, err := config.Load(fs)
	if err != nil {
		return err
	}
	if cfg.MasterKey == "" {
		return fmt.Errorf("master_key is required")
	}

	enc, err := envelope.New(cfg.MasterKey)
	if err != nil {
		return fmt.Errorf("failed to init envelope encrypter: %w", err)
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
	secSvc := service.NewSecretsService(storage, enc)

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

		// Disable HTTP/2 for compatibility with self-signed TLS certificates.
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}

	errCh := make(chan error, 1)

	go func() {
		logger.Log.Info("server started", zap.String("addr", cfg.Addr))

		var err error

		if cfg.TLSCert != "" && cfg.TLSKey != "" {
			logger.Log.Info("starting HTTPS server")
			err = srv.ListenAndServeTLS(cfg.TLSCert, cfg.TLSKey)
		} else {
			logger.Log.Warn("starting HTTP server (TLS disabled)")
			err = srv.ListenAndServe()
		}

		if err != nil && !errors.Is(err, http.ErrServerClosed) {
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
