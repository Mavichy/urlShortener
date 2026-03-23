package main

import (
	"context"
	"github.com/Mavichy/urlShortener/internal/app"
	"github.com/Mavichy/urlShortener/internal/config"
	"log"
	"os/signal"
	"syscall"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := app.Run(ctx, cfg); err != nil {
		log.Fatalf("run app: %v", err)
	}
}
