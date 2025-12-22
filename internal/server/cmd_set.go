package server

import (
	"context"

	"github.com/tidwall/redcon"
)

func (server *Server) handleSet(ctx context.Context, conn redcon.Conn, cmd redcon.Command) {
	if isInvalidManyArgs(cmd) {
		conn.WriteError("ERR wrong number of arguments for 'set' command")
		return
	}

	key := cmd.Args[firstArg]
	value := cmd.Args[secondArg]

	err := server.storage.SetWithContext(ctx, key, value)
	if HasError(err) {
		if isContextCanceled(err) {
			conn.WriteError("ERR operation canceled")
			return
		}
		conn.WriteError(err.Error())
		return
	}

	conn.WriteString("OK")
}
