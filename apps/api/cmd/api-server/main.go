package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/dendianugerah/velld/internal/app"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = ":8080"
	}

	if err := app.Run(ctx, addr); err != nil {
		log.Fatalf("server exited with error: %v", err)
	}
}
