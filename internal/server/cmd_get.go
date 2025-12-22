package server

import "github.com/tidwall/redcon"

func (server *Server) handleGet(conn redcon.Conn, cmd redcon.Command) {
	if isInvalidTwoArgs(cmd) {
		conn.WriteError("ERR wrong number of arguments for 'get' command")
		return
	}

	key := cmd.Args[firstArg]
	value, err := server.storage.Get(key)

	if HasError(err) {
		if isKeyNotFoundError(err) {
			conn.WriteNull()
			return
		}

		conn.WriteError(err.Error())
		return
	}

	conn.WriteBulk(value)
}
