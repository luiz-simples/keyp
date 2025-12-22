package server

import (
	"strconv"

	"github.com/tidwall/redcon"
)

func (server *Server) handleExpireAt(conn redcon.Conn, cmd redcon.Command) {
	if isInvalidManyArgs(cmd) {
		conn.WriteError("ERR wrong number of arguments for 'expireat' command")
		return
	}

	key := cmd.Args[firstArg]
	timestampStr := string(cmd.Args[secondArg])

	timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if hasError(err) {
		conn.WriteError("ERR value is not an integer or out of range")
		return
	}

	ttlManager := server.storage.GetTTLManager()
	result, err := ttlManager.SetExpireAt(key, timestamp)
	if hasError(err) {
		conn.WriteError("ERR " + err.Error())
		return
	}

	conn.WriteInt(result)
}
