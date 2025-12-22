package server

import (
	"errors"
	"time"

	"github.com/tidwall/redcon"

	"github.com/luiz-simples/keyp.git/internal/storage"
)

const (
	noArgs      = 0
	singleArg   = 1
	twoArgs     = 2
	threeArgs   = 3
	minMultArgs = 2
	firstArg    = 1
	secondArg   = 2
)

func hasError(err error) bool {
	return err != nil
}

func emptyArgs(cmd redcon.Command) bool {
	return len(cmd.Args) == noArgs
}

func hasSingleArg(cmd redcon.Command) bool {
	return len(cmd.Args) == singleArg
}

func hasTwoArgs(cmd redcon.Command) bool {
	return len(cmd.Args) == twoArgs
}

func isInvalidManyArgs(cmd redcon.Command) bool {
	return len(cmd.Args) != threeArgs
}

func isInvalidTwoArgs(cmd redcon.Command) bool {
	return len(cmd.Args) != twoArgs
}

func isInvalidMultArgs(cmd redcon.Command) bool {
	return len(cmd.Args) < minMultArgs
}

func handlePing(conn redcon.Conn, cmd redcon.Command) {
	if hasSingleArg(cmd) {
		conn.WriteString("PONG")
		return
	}

	if hasTwoArgs(cmd) {
		conn.WriteBulk(cmd.Args[firstArg])
		return
	}

	conn.WriteError("ERR wrong number of arguments for 'ping' command")
}

func isKeyNotFoundError(err error) bool {
	return errors.Is(err, storage.ErrKeyNotFound)
}
func getCleanupInterval() time.Duration {
	return 60 * time.Second
}
