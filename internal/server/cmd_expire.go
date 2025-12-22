package server

import (
	"strconv"

	"github.com/tidwall/redcon"
)

func (server *Server) handleExpire(conn redcon.Conn, cmd redcon.Command) {
	if isInvalidManyArgs(cmd) {
		conn.WriteError("ERR wrong number of arguments for 'expire' command")
		return
	}

	key := cmd.Args[firstArg]
	secondsStr := string(cmd.Args[secondArg])

	seconds, err := strconv.ParseInt(secondsStr, 10, 64)
	if HasError(err) {
		conn.WriteError("ERR value is not an integer or out of range")
		return
	}

	ttlManager := server.storage.GetTTLManager()
	result, err := ttlManager.SetExpire(key, seconds)
	if HasError(err) {
		conn.WriteError("ERR " + err.Error())
		return
	}

	conn.WriteInt(result)
}
