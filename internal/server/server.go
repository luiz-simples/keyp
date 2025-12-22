package server

import (
	"errors"
	"log"
	"strings"

	"github.com/tidwall/redcon"

	"github.com/luiz-simples/keyp.git/internal/storage"
)

type Server struct {
	addr    string
	server  *redcon.Server
	storage *storage.LMDBStorage
}

func New(addr, dataDir string) (*Server, error) {
	storage, err := storage.NewLMDBStorage(dataDir)
	if err != nil {
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
		_ = server.storage.Close()
	}
	return server.server.Close()
}

func (server *Server) handleConnect(conn redcon.Conn) bool {
	log.Printf("Client connected: %s", conn.RemoteAddr())
	return true
}

func (server *Server) handleClose(conn redcon.Conn, err error) {
	log.Printf("Client disconnected: %s", conn.RemoteAddr())
	if err != nil {
		log.Printf("Connection error: %v", err)
	}
}

func (server *Server) handleCommand(conn redcon.Conn, cmd redcon.Command) {
	if len(cmd.Args) == 0 {
		conn.WriteError("ERR empty command")
		return
	}

	command := strings.ToUpper(string(cmd.Args[0]))

	handlers := map[string]func(redcon.Conn, redcon.Command){
		"PING": handlePing,
		"SET":  server.handleSet,
		"GET":  server.handleGet,
		"DEL":  server.handleDel,
	}

	handler, exists := handlers[command]
	if !exists {
		conn.WriteError("ERR unknown command '" + command + "'")
		return
	}

	handler(conn, cmd)
}

func handlePing(conn redcon.Conn, cmd redcon.Command) {
	if len(cmd.Args) == 1 {
		conn.WriteString("PONG")
		return
	}
	if len(cmd.Args) == 2 {
		conn.WriteBulk(cmd.Args[1])
		return
	}
	conn.WriteError("ERR wrong number of arguments for 'ping' command")
}

func (server *Server) handleSet(conn redcon.Conn, cmd redcon.Command) {
	if len(cmd.Args) != 3 {
		conn.WriteError("ERR wrong number of arguments for 'set' command")
		return
	}

	key := cmd.Args[1]
	value := cmd.Args[2]

	err := server.storage.Set(key, value)
	if err != nil {
		conn.WriteError("ERR " + err.Error())
		return
	}

	conn.WriteString("OK")
}

func (server *Server) handleGet(conn redcon.Conn, cmd redcon.Command) {
	if len(cmd.Args) != 2 {
		conn.WriteError("ERR wrong number of arguments for 'get' command")
		return
	}

	key := cmd.Args[1]

	value, err := server.storage.Get(key)
	if err != nil {
		if errors.Is(err, storage.ErrKeyNotFound) {
			conn.WriteNull()
			return
		}
		conn.WriteError("ERR " + err.Error())
		return
	}

	conn.WriteBulk(value)
}

func (server *Server) handleDel(conn redcon.Conn, cmd redcon.Command) {
	if len(cmd.Args) < 2 {
		conn.WriteError("ERR wrong number of arguments for 'del' command")
		return
	}

	keys := cmd.Args[1:]
	deleted, err := server.storage.Del(keys...)
	if err != nil {
		conn.WriteError("ERR " + err.Error())
		return
	}

	conn.WriteInt(deleted)
}
