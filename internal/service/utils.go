package service

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/luiz-simples/keyp.git/internal/domain"
)

func hasError(err error) bool {
	return err != nil
}

func noError(err error) bool {
	return err == nil
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

	res.Response = formatInt64(result)
	return res
}

func formatInt64(value int64) []byte {
	return []byte(strconv.FormatInt(value, 10))
}

func formatUint32(value uint32) []byte {
	return []byte(strconv.FormatUint(uint64(value), 10))
}

func formatBool(value bool) []byte {
	if value {
		return []byte("1")
	}
	return []byte("0")
}

func formatArray(items [][]byte) []byte {
	if len(items) == 0 {
		return []byte("*0\r\n")
	}

	result := []byte("*" + strconv.Itoa(len(items)) + "\r\n")
	for _, item := range items {
		result = append(result, []byte("$"+strconv.Itoa(len(item))+"\r\n")...)
		result = append(result, item...)
		result = append(result, []byte("\r\n")...)
	}
	return result
}
