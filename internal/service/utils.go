package service

import (
	"context"
	"encoding/binary"
	"errors"
	"strings"

	"github.com/luiz-simples/keyp.git/internal/domain"
)

var (
	PONG   []byte = []byte("PONG")
	QUEUED []byte = []byte("QUEUED")

	EMPTY error = errors.New("ERR empty command")

	ErrInvalidInteger = errors.New("ERR value is not an integer or out of range")
	ErrInvalidFloat   = errors.New("ERR value is not a valid float")
)

const (
	PING    string = "PING"
	MULTI   string = "MULTI"
	EXEC    string = "EXEC"
	DISCARD string = "DISCARD"

	noArgs    = 0
	firstArg  = 1
	secondArg = 2
	thirdArg  = 3
)

func hasError(err error) bool {
	return err != nil
}

func emptyArgs(args Args) bool {
	return len(args) == noArgs
}

func isKeyNotFoundError(err error) bool {
	return hasError(err) && strings.Contains(err.Error(), "key not found")
}

func newInvalidArgsError(commandName string) error {
	return errors.New("ERR wrong number of arguments for '" + commandName + "' command")
}

func normalizeCommandName(name string) string {
	return strings.ToUpper(strings.TrimSpace(name))
}

func isContextCanceled(err error) bool {
	return hasError(err) && (errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded))
}

func isValid(validation *domain.Validation, cmdName string, argCount int) error {
	if argCount < validation.MinArgs {
		return newInvalidArgsError(cmdName)
	}

	if validation.MaxArgs > 0 && argCount > validation.MaxArgs {
		return newInvalidArgsError(cmdName)
	}

	return nil
}

func encodeArray(items [][]byte) []byte {
	if len(items) == 0 {
		response := make([]byte, 8)
		binary.LittleEndian.PutUint64(response, 0)
		return response
	}

	response := make([]byte, 8)
	binary.LittleEndian.PutUint64(response, uint64(len(items)))

	for _, item := range items {
		response = append(response, make([]byte, 4)...)
		binary.LittleEndian.PutUint32(response[len(response)-4:], uint32(len(item)))
		response = append(response, item...)
	}

	return response
}
