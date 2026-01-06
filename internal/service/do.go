package service

import (
	"errors"

	"github.com/luiz-simples/keyp.git/internal/domain"
)

func (handler *Handler) do(args Args) *Result {
	if len(args) < 2 {
		return domain.NewResult().SetEmpty()
	}

	commandName := normalizeCommandName(string(args[domain.CommandArg]))
	commandArgs := append([][]byte{[]byte(commandName)}, args[domain.FirstArg:]...)
	command, exists := handler.commands[commandName]

	if !exists {
		result := domain.NewResult()
		result.Error = errors.New("ERR unknown command '" + commandName + "'")
		return result
	}

	return command(commandArgs)
}
