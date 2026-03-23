package app

import (
	"context"
	"errors"
	"fmt"
	"github.com/Mavichy/urlShortener/internal/config"
	"github.com/Mavichy/urlShortener/internal/httpserver"
	"github.com/Mavichy/urlShortener/internal/service"
	"github.com/Mavichy/urlShortener/internal/shortcode"
	"github.com/Mavichy/urlShortener/internal/storage"
	memorystore "github.com/Mavichy/urlShortener/internal/storage/memory"
	postgresstore "github.com/Mavichy/urlShortener/internal/storage/postgres"
	"log"
	"net/http"
)

func Run(ctx context.Context, cfg config.Config) error {
	store, err := buildStorage(ctx, cfg)
	if err != nil {
		return err
	}
	defer func() {
		if err := store.Close(); err != nil {
			log.Printf("close storage: %v", err)
		}
	}()

	svc := service.New(store, shortcode.NewGenerator())

	srv := httpserver.New(httpserver.ServerConfig{
		Addr:           cfg.HTTPAddr,
		BaseURL:        cfg.BaseURL,
		RequestTimeout: cfg.RequestTimeout,
		ReadTimeout:    cfg.ReadTimeout,
		WriteTimeout:   cfg.WriteTimeout,
		IdleTimeout:    cfg.IdleTimeout,
	}, svc)

	errCh := make(chan error, 1)
	go func() {
		log.Printf("starting http server on %s using storage=%s", cfg.HTTPAddr, cfg.StorageType)
		errCh <- srv.Start()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()

		log.Printf("shutting down http server")
		return srv.Shutdown(shutdownCtx)

	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}

func buildStorage(ctx context.Context, cfg config.Config) (storage.Storage, error) {
	switch cfg.StorageType {
	case config.StorageMemory:
		return memorystore.New(), nil

	case config.StoragePostgres:
		store, err := postgresstore.New(ctx, cfg.PostgresDSN)
		if err != nil {
			return nil, fmt.Errorf("create postgres storage: %w", err)
		}
		return store, nil

	default:
		return nil, fmt.Errorf("unsupported storage type %q", cfg.StorageType)
	}
}
