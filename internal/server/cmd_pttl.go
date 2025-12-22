package server

import "github.com/tidwall/redcon"

func (server *Server) handlePTTL(conn redcon.Conn, cmd redcon.Command) {
	if isInvalidTwoArgs(cmd) {
		conn.WriteError("ERR wrong number of arguments for 'pttl' command")
		return
	}

	key := cmd.Args[firstArg]

	ttlManager := server.storage.GetTTLManager()
	pttl, err := ttlManager.GetPTTL(key)
	if HasError(err) {
		conn.WriteError(err.Error())
		return
	}

	conn.WriteInt64(pttl)
}
