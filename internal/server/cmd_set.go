package server

import "github.com/tidwall/redcon"

func (server *Server) handleSet(conn redcon.Conn, cmd redcon.Command) {
	if isInvalidManyArgs(cmd) {
		conn.WriteError("ERR wrong number of arguments for 'set' command")
		return
	}

	key := cmd.Args[firstArg]
	value := cmd.Args[secondArg]

	err := server.storage.Set(key, value)
	if HasError(err) {
		conn.WriteError(err.Error())
		return
	}

	conn.WriteString("OK")
}
