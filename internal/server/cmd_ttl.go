package server

import "github.com/tidwall/redcon"

func (server *Server) handleTTL(conn redcon.Conn, cmd redcon.Command) {
	if isInvalidTwoArgs(cmd) {
		conn.WriteError("ERR wrong number of arguments for 'ttl' command")
		return
	}

	key := cmd.Args[firstArg]

	ttlManager := server.storage.GetTTLManager()
	ttl, err := ttlManager.GetTTL(key)
	if HasError(err) {
		conn.WriteError(err.Error())
		return
	}

	conn.WriteInt64(ttl)
}
