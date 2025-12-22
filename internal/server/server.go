package server

import (
	"context"
	"strings"
	"time"

	"github.com/tidwall/redcon"

	"github.com/luiz-simples/keyp.git/internal/logger"
	"github.com/luiz-simples/keyp.git/internal/storage"
)

type Server struct {
	addr           string
	server         *redcon.Server
	storage        *storage.LMDBStorage
	cleanupCtx     context.Context
	cleanupCancel  context.CancelFunc
	cleanupStopped chan struct{}
}

func New(addr, dataDir string) (*Server, error) {
	storage, err := storage.NewLMDBStorage(dataDir)
	if hasError(err) {
		return nil, err
	}

	ttlManager := storage.GetTTLManager()
	err = ttlManager.RestoreTTL()
	if hasError(err) {
		logger.Error("TTL restoration failed", "error", err)
	}

	cleanupCtx, cleanupCancel := context.WithCancel(context.Background())

	server := &Server{
		addr:           addr,
		storage:        storage,
		cleanupCtx:     cleanupCtx,
		cleanupCancel:  cleanupCancel,
		cleanupStopped: make(chan struct{}),
	}

	server.server = redcon.NewServer(addr,
		server.handleCommand,
		server.handleConnect,
		server.handleClose,
	)

	return server, nil
}

func (server *Server) ListenAndServe() error {
	go server.startBackgroundCleanup()
	return server.server.ListenAndServe()
}

func (server *Server) Close() error {
	server.cleanupCancel()
	<-server.cleanupStopped

	if server.storage != nil {
		server.storage.Close()
	}
	return server.server.Close()
}

func (server *Server) handleConnect(conn redcon.Conn) bool {
	logger.Debug("Client connected", "addr", conn.RemoteAddr())
	return true
}

func (server *Server) handleClose(conn redcon.Conn, err error) {
	logger.Debug("Client disconnected", "addr", conn.RemoteAddr())
	if hasError(err) {
		logger.Error("Connection error", "error", err)
	}
}

func (server *Server) handleCommand(conn redcon.Conn, cmd redcon.Command) {
	if emptyArgs(cmd) {
		conn.WriteError("ERR empty command")
		return
	}

	command := strings.ToUpper(string(cmd.Args[0]))

	handlers := map[string]func(redcon.Conn, redcon.Command){
		"PING":     handlePing,
		"SET":      server.handleSet,
		"GET":      server.handleGet,
		"DEL":      server.handleDel,
		"EXPIRE":   server.handleExpire,
		"EXPIREAT": server.handleExpireAt,
		"TTL":      server.handleTTL,
		"PTTL":     server.handlePTTL,
		"PERSIST":  server.handlePersist,
	}

	handler, exists := handlers[command]
	if !exists {
		conn.WriteError("ERR unknown command '" + command + "'")
		return
	}

	handler(conn, cmd)
}
func (server *Server) startBackgroundCleanup() {
	defer close(server.cleanupStopped)

	ticker := time.NewTicker(getCleanupInterval())
	defer ticker.Stop()

	for {
		select {
		case <-server.cleanupCtx.Done():
			return
		case <-ticker.C:
			server.performCleanup()
		}
	}
}

func (server *Server) performCleanup() {
	ttlManager := server.storage.GetTTLManager()
	err := ttlManager.CleanupExpired()
	if hasError(err) {
		logger.Error("Background cleanup failed", "error", err)
		return
	}

	logger.Debug("Background cleanup completed")
}
