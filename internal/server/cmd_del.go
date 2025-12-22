package server

import "github.com/tidwall/redcon"

func (server *Server) handleDel(conn redcon.Conn, cmd redcon.Command) {
	if isInvalidMultArgs(cmd) {
		conn.WriteError("ERR wrong number of arguments for 'del' command")
		return
	}

	keys := cmd.Args[firstArg:]
	deleted, err := server.storage.Del(keys...)

	if hasError(err) {
		conn.WriteError("ERR " + err.Error())
		return
	}

	conn.WriteInt(deleted)
}
