package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/luiz-simples/keyp.git/internal/server"
)

const (
	defaultPort    = "6380"
	defaultHost    = "localhost"
	defaultDataDir = "./data"
)

func main() {
	var (
		port    = flag.String("port", defaultPort, "Port to listen on")
		host    = flag.String("host", defaultHost, "Host to bind to")
		dataDir = flag.String("data-dir", defaultDataDir, "Data directory for LMDB")
	)
	flag.Parse()

	addr := fmt.Sprintf("%s:%s", *host, *port)

	srv, err := server.New(addr, *dataDir)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Println("Shutting down server...")
		_ = srv.Close()
		os.Exit(0)
	}()

	log.Printf("Keyp server starting on %s", addr)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
