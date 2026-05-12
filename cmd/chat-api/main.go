package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/jandreyh/chat-api/internal/config"
	"github.com/jandreyh/chat-api/internal/server"
)

func main() {
	cfg := config.Load()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := server.New(cfg).Run(ctx); err != nil {
		log.Fatalf("chat-api: %v", err)
	}
}
