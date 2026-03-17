package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/uucz/llmview/internal/server"
)

var version = "dev"

func main() {
	port := flag.Int("port", 4700, "port to listen on")
	dbPath := flag.String("db", "", "SQLite database path (default: ~/.llmview/llmview.db)")
	budget := flag.Float64("budget", 0, "max session cost in USD (0 = unlimited)")
	showVersion := flag.Bool("version", false, "show version")
	flag.Parse()

	if *showVersion {
		fmt.Printf("llmview %s\n", version)
		os.Exit(0)
	}

	sessionID := fmt.Sprintf("s_%d", time.Now().UnixMilli())

	srv, err := server.New(server.Config{
		Port:      *port,
		DBPath:    *dbPath,
		SessionID: sessionID,
		Budget:    *budget,
	})
	if err != nil {
		log.Fatalf("failed to start: %v", err)
	}

	// Graceful shutdown
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		log.Println("shutting down...")
		srv.Close()
		os.Exit(0)
	}()

	if err := srv.Start(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
