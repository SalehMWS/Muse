package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/SalehMWS/Muse/internal/shared/bootstrap"
	"github.com/SalehMWS/Muse/internal/shared/config"
)

func main() {
	healthcheck := flag.Bool("healthcheck", false, "probe the local /health/live endpoint and exit (used by Docker HEALTHCHECK)")
	flag.Parse()

	if *healthcheck {
		os.Exit(runHealthcheck())
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx); err != nil {
		log.Fatal(err)
	}
}

func runHealthcheck() int {
	cfg, err := config.Load()
	if err != nil {
		return 1
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	url := fmt.Sprintf("http://127.0.0.1:%d/health/live", cfg.HTTP.Port)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 1
	}

	client := http.Client{Timeout: 2 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 1
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return 1
	}
	return 0
}

func run(ctx context.Context) error {
	container, err := bootstrap.New(ctx)
	if err != nil {
		return fmt.Errorf("start api: %w", err)
	}

	serveErrors := make(chan error, 1)
	go func() {
		addr := fmt.Sprintf(":%d", container.Config.HTTP.Port)
		container.Logger.Info("http server starting",
			zap.String("module", "api"),
			zap.String("addr", addr),
		)

		if err := container.App.Listen(addr); err != nil {
			serveErrors <- err
		}
	}()

	select {
	case err := <-serveErrors:
		return fmt.Errorf("http server: %w", err)
	case <-ctx.Done():
		container.Logger.Info("shutdown signal received", zap.String("module", "api"))
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), container.Config.HTTP.ShutdownTimeout)
	defer cancel()

	if err := container.App.ShutdownWithContext(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		container.Logger.Error("http server shutdown", zap.Error(err))
	}

	if err := container.Shutdown(shutdownCtx); err != nil {
		container.Logger.Error("dependency shutdown", zap.Error(err))
	}

	container.Logger.Info("shutdown complete", zap.String("module", "api"))
	time.Sleep(50 * time.Millisecond)

	return nil
}
