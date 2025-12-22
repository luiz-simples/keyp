package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/luiz-simples/keyp.git/internal/logger"
	"github.com/luiz-simples/keyp.git/internal/server"
)

const (
	defaultPort      = "6380"
	defaultHost      = "localhost"
	defaultDataDir   = "./data"
	signalBufferSize = 1
	successExitCode  = 0
)

func hasError(err error) bool {
	return err != nil
}

func main() {
	var (
		port    = flag.String("port", defaultPort, "Port to listen on")
		host    = flag.String("host", defaultHost, "Host to bind to")
		dataDir = flag.String("data-dir", defaultDataDir, "Data directory for LMDB")
	)
	flag.Parse()

	addr := fmt.Sprintf("%s:%s", *host, *port)

	srv, err := server.New(addr, *dataDir)
	if hasError(err) {
		logger.Fatal("Failed to create server", "error", err)
	}

	c := make(chan os.Signal, signalBufferSize)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		logger.Info("Shutting down server...")
		srv.Close()
		os.Exit(successExitCode)
	}()

	logger.Info("Keyp server starting", "addr", addr)

	if err := srv.ListenAndServe(); hasError(err) {
		logger.Fatal("Server error", "error", err)
	}
}
