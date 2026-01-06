package app

import (
	"context"
	"sync"
	"time"

	"github.com/tidwall/redcon"

	"github.com/luiz-simples/keyp.git/internal/domain"
)

type (
	Server struct {
		rcon     *redcon.Server
		handlers map[int64]domain.Dispatcher
		poolHdlr domain.Logicaler
		mutex    sync.RWMutex
	}

	Config struct {
		Address string
		DataDir string
	}
)

func NewServer(pool domain.Logicaler) *Server {
	return &Server{
		handlers: make(map[int64]domain.Dispatcher),
		poolHdlr: pool,
	}
}

func (server *Server) Start(config Config) error {
	server.rcon = redcon.NewServer(
		config.Address,
		server.OnHandler,
		server.OnAccept,
		server.OnClosed,
	)

	return server.rcon.ListenAndServe()
}

func (server *Server) OnHandler(conn redcon.Conn, cmd redcon.Command) {
	ctx, ok := conn.Context().(context.Context)
	if !ok {
		conn.WriteError("ERR invalid connection context")
		return
	}

	connID, ok := ctx.Value(domain.ID).(int64)
	if !ok {
		conn.WriteError("ERR invalid connection ID")
		return
	}

	var handler domain.Dispatcher

	for range 10 {
		handler = server.getHandler(connID)

		if handler != nil {
			break
		}

		time.Sleep(time.Millisecond)
	}

	if handler == nil {
		conn.WriteError("ERR connection not found")
		return
	}

	results := handler.Apply(ctx, cmd.Args)

	for _, item := range results {
		if hasError(item.Error) {
			conn.WriteError(item.Error.Error())
			continue
		}

		if item.Response == nil {
			conn.WriteNull()
			continue
		}

		if hasResponse(item.Response) {
			if isArrayResponse(item.Response) {
				conn.WriteRaw(item.Response)
				continue
			}

			conn.WriteBulk(item.Response)
			continue
		}

		conn.WriteNull()
	}
}

func (server *Server) getHandler(connID int64) domain.Dispatcher {
	server.mutex.RLock()
	defer server.mutex.RUnlock()
	return server.handlers[connID]
}

func (server *Server) OnAccept(conn redcon.Conn) bool {
	connID := generateConnectionID()
	ctx := context.WithValue(context.Background(), domain.ID, connID)
	ctx = context.WithValue(ctx, domain.DB, uint8(0))

	conn.SetContext(ctx)

	server.mutex.Lock()
	server.handlers[connID] = server.poolHdlr.Get(ctx)
	server.mutex.Unlock()

	return true
}

func (server *Server) OnClosed(conn redcon.Conn, err error) {
	ctx, ok := conn.Context().(context.Context)
	if !ok {
		return
	}

	connID, ok := ctx.Value(domain.ID).(int64)
	if !ok {
		return
	}

	server.mutex.Lock()
	defer server.mutex.Unlock()

	if handlerExists(server.handlers, connID) {
		dispatcher := server.handlers[connID]
		delete(server.handlers, connID)
		server.poolHdlr.Free(dispatcher)
	}
}

func (server *Server) Close() {
	server.mutex.Lock()
	defer server.mutex.Unlock()

	for _, handler := range server.handlers {
		handler.Clear()
	}

	server.handlers = make(map[int64]domain.Dispatcher)

	if server.rcon != nil {
		server.rcon.Close()
		server.rcon = nil
	}
}
