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
	registry       *CommandRegistry
}

func New(addr, dataDir string) (*Server, error) {
	storage, err := storage.NewLMDBStorage(dataDir)
	if HasError(err) {
		return nil, err
	}

	ttlManager := storage.GetTTLManager()
	err = ttlManager.RestoreTTL()
	if HasError(err) {
		logger.Error("TTL restoration failed", "error", err)
	}

	cleanupCtx, cleanupCancel := context.WithCancel(context.Background())

	server := &Server{
		addr:           addr,
		storage:        storage,
		cleanupCtx:     cleanupCtx,
		cleanupCancel:  cleanupCancel,
		cleanupStopped: make(chan struct{}),
		registry:       NewCommandRegistry(),
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
	if HasError(err) {
		logger.Error("Connection error", "error", err)
	}
}

func (server *Server) handleCommand(conn redcon.Conn, cmd redcon.Command) {
	if emptyArgs(cmd) {
		conn.WriteError("ERR empty command")
		return
	}

	commandName := strings.ToUpper(string(cmd.Args[0]))

	metadata, exists := server.registry.GetCommand(commandName)
	if !exists {
		conn.WriteError("ERR unknown command '" + commandName + "'")
		return
	}

	err := server.registry.ValidateCommand(cmd, metadata)
	if HasError(err) {
		conn.WriteError(err.Error())
		return
	}

	ctx := context.Background()
	server.executeCommand(ctx, conn, cmd, metadata)
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
	if HasError(err) {
		logger.Error("Background cleanup failed", "error", err)
		return
	}

	logger.Debug("Background cleanup completed")
}
func (server *Server) executeCommand(ctx context.Context, conn redcon.Conn, cmd redcon.Command, metadata *CommandMetadata) {
	if metadata.Handler != nil {
		metadata.Handler(ctx, conn, cmd)
		return
	}

	commandName := metadata.Name
	contextHandlers := map[string]func(context.Context, redcon.Conn, redcon.Command){
		"SET": server.handleSet,
	}

	if handler, exists := contextHandlers[commandName]; exists {
		handler(ctx, conn, cmd)
		return
	}

	legacyHandlers := map[string]func(redcon.Conn, redcon.Command){
		"GET":      server.handleGet,
		"DEL":      server.handleDel,
		"EXPIRE":   server.handleExpire,
		"EXPIREAT": server.handleExpireAt,
		"TTL":      server.handleTTL,
		"PTTL":     server.handlePTTL,
		"PERSIST":  server.handlePersist,
	}

	if handler, exists := legacyHandlers[commandName]; exists {
		handler(conn, cmd)
		return
	}

	conn.WriteError("ERR command '" + commandName + "' not implemented")
}
