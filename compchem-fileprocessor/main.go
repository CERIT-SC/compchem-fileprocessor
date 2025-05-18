package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"time"

	"fi.muni.cz/invenio-file-processor/v2/config"
	"fi.muni.cz/invenio-file-processor/v2/db"
	"fi.muni.cz/invenio-file-processor/v2/routes"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// credit to: https://grafana.com/blog/2024/02/09/how-i-write-http-services-in-go-after-13-years/
func NewServer(
	ctx context.Context,
	logger *zap.Logger,
	pool *pgxpool.Pool,
	config *config.Config,
) http.Handler {
	mux := http.NewServeMux()

	routes.AddRoutes(ctx, logger, mux, config)

	return mux
}

func getConfig(logger *zap.Logger) (*config.Config, error) {
	wd, err := os.Getwd()
	if err != nil {
		logger.Error("Could not get workdir", zap.Error(err))
		return nil, err
	}

	config, err := config.LoadConfig(logger, wd)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func run(
	ctx context.Context,
	logger *zap.Logger,
	config *config.Config,
	ready chan<- struct{},
) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	pool, err := db.CreatePgPool(ctx, logger, &config.Postgres, config.Migrations)
	if err != nil {
		logger.Error("Error initializing db connection pool")
		return err
	}
	defer pool.Close()

	srv := NewServer(ctx, logger, pool, config)
	httpServer := &http.Server{
		Addr:    net.JoinHostPort(config.Server.Host, strconv.Itoa(config.Server.Port)),
		Handler: srv,
	}

	go func() {
		logger.Info(
			"Listening on",
			zap.String("host", config.Server.Host),
			zap.Int("port", config.Server.Port),
		)

		close(ready)

		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Error listening", zap.Error(err))
		}
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		shutdownCtx := context.Background()
		shutdownCtx, cancel := context.WithTimeout(shutdownCtx, 10*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			logger.Error("Error during shutdown", zap.Error(err))
		}
	}()
	wg.Wait()
	return nil
}

func main() {
	ctx := context.Background()
	ready := make(chan struct{})

	logger, _ := zap.NewProduction()
	defer logger.Sync()

	config, err := getConfig(logger)
	if err != nil {
		os.Exit(1)
	}

	if err := run(ctx, logger, config, ready); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
