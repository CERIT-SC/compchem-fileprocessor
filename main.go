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

	"fi.muni.cz/upload-processor/v2/config"
	"go.uber.org/zap"
)

func NewServer(ctx context.Context, logger *zap.Logger) http.Handler {
	mux := http.NewServeMux()

	return mux
}

func run(ctx context.Context, args []string) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	config, err := config.LoadConfig(logger, args[0])
	if err != nil {
		return err
	}

	srv := NewServer(ctx, logger)
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
	if err := run(ctx, os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
