package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"report-service-m/internal/config"
	"report-service-m/internal/httpserver"
)

func main() {
	cfg := config.Load()
	app := httpserver.New(cfg)

	serverErr := make(chan error, 1)
	go func() {
		log.Printf("%s listening on %s", cfg.AppName, cfg.Addr())
		serverErr <- app.Listen(cfg.Addr())
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		if err != nil {
			log.Fatal(err)
		}
	case sig := <-shutdown:
		log.Printf("received %s, shutting down", sig)
		if err := app.ShutdownWithTimeout(10 * time.Second); err != nil {
			log.Printf("graceful shutdown failed: %v", err)
		}
		if err := <-serverErr; err != nil {
			log.Printf("server stopped with error: %v", err)
		}
	}
}
