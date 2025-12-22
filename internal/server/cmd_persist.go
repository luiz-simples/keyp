package server

import "github.com/tidwall/redcon"

func (server *Server) handlePersist(conn redcon.Conn, cmd redcon.Command) {
	if isInvalidTwoArgs(cmd) {
		conn.WriteError("ERR wrong number of arguments for 'persist' command")
		return
	}

	key := cmd.Args[firstArg]

	ttlManager := server.storage.GetTTLManager()
	result, err := ttlManager.Persist(key)
	if HasError(err) {
		conn.WriteError(err.Error())
		return
	}

	conn.WriteInt(result)
}
