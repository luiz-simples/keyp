package app

import (
	"context"

	"github.com/tidwall/redcon"

	"github.com/luiz-simples/keyp.git/internal/domain"
)

type (
	Server struct {
		contexts map[int64]func()
		handlers map[int64]domain.Dispatcher
		poolHdlr domain.Logicaler
	}

	Config struct {
		Address string
		DataDir string
	}
)

func NewServer(pool domain.Logicaler) *Server {
	return &Server{
		contexts: make(map[int64]func()),
		handlers: make(map[int64]domain.Dispatcher),
		poolHdlr: pool,
	}
}

func (server *Server) Start(config Config) error {
	return redcon.ListenAndServe(
		config.Address,
		server.OnHandler,
		server.OnAccept,
		server.OnClosed,
	)
}

func (server *Server) OnHandler(conn redcon.Conn, cmd redcon.Command) {
	ctx, _ := conn.Context().(context.Context)
	connID, _ := ctx.Value(domain.ID).(int64)
	results := server.handlers[connID].Apply(ctx, cmd.Args)

	for _, item := range results {
		if hasError(item.Error) {
			conn.WriteError(item.Error.Error())
			continue
		}

		if hasResponse(item.Response) {
			conn.WriteBulk(item.Response)
			continue
		}

		conn.WriteNull()
	}
}

func (server *Server) OnAccept(conn redcon.Conn) bool {
	connID := generateConnectionID()
	ctx, cancel := context.WithCancel(context.WithValue(context.Background(), domain.ID, connID))

	conn.SetContext(ctx)
	server.contexts[connID] = cancel
	server.handlers[connID] = server.poolHdlr.Get(ctx)

	return true
}

func (server *Server) OnClosed(conn redcon.Conn, err error) {
	ctx, _ := conn.Context().(context.Context)
	connID, _ := ctx.Value(domain.ID).(int64)

	if contextExists(server.contexts, connID) {
		cancel := server.contexts[connID]
		delete(server.contexts, connID)
		cancel()
	}

	if handlerExists(server.handlers, connID) {
		dispatcher := server.handlers[connID]
		delete(server.handlers, connID)
		server.poolHdlr.Free(dispatcher)
	}
}

func (server *Server) Close() {
	for _, cancel := range server.contexts {
		cancel()
	}
}
