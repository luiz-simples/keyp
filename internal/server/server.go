package server

import (
	"strings"

	"github.com/tidwall/redcon"

	"github.com/luiz-simples/keyp.git/internal/logger"
	"github.com/luiz-simples/keyp.git/internal/storage"
)

type Server struct {
	addr    string
	server  *redcon.Server
	storage *storage.LMDBStorage
}

func New(addr, dataDir string) (*Server, error) {
	storage, err := storage.NewLMDBStorage(dataDir)
	if hasError(err) {
		return nil, err
	}

	server := &Server{
		addr:    addr,
		storage: storage,
	}

	server.server = redcon.NewServer(addr,
		server.handleCommand,
		server.handleConnect,
		server.handleClose,
	)

	return server, nil
}

func (server *Server) ListenAndServe() error {
	return server.server.ListenAndServe()
}

func (server *Server) Close() error {
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
