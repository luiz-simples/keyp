package service

import (
	"context"
	"encoding/binary"
	"errors"
	"strconv"
	"strings"

	"github.com/luiz-simples/keyp.git/internal/domain"
)

func hasError(err error) bool {
	return err != nil
}

func emptyArgs(args Args) bool {
	return len(args) == domain.EmptyArgs
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

func processIntegerModification(args Args, storageMethod func(context.Context, []byte, int64) (int64, error), handler *Handler) *Result {
	res := domain.NewResult()
	key := args[domain.FirstArg]
	valueStr := string(args[domain.SecondArg])

	value, err := strconv.ParseInt(valueStr, 10, 64)
	if hasError(err) {
		res.Error = domain.ErrInvalidInteger
		return res
	}

	result, err := storageMethod(handler.context, key, value)
	if hasError(err) {
		res.Error = err
		return res
	}

	res.Response = make([]byte, 8)
	binary.LittleEndian.PutUint64(res.Response, uint64(result))

	return res
}
