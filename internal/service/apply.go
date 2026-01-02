package service

import (
	"context"
	"errors"

	"github.com/luiz-simples/keyp.git/internal/domain"
)

func (handler *Handler) Apply(ctx context.Context, args Args) Results {
	if emptyArgs(args) {
		return Results{{Error: domain.ErrEmpty}}
	}

	cmdName := normalizeCommandName(string(args[0]))

	if cmdName == domain.MULTI {
		handler.multEnabled = true
		return Results{{Response: OK}}
	}

	if cmdName == domain.DISCARD {
		handler.multEnabled = false
		handler.multArgs = handler.multArgs[:0]
		return Results{{Response: OK}}
	}

	if cmdName == domain.EXEC {
		return handler.exec()
	}

	validation, exists := handler.validations[cmdName]

	if !exists {
		return Results{{Error: errors.New("ERR unknown command '" + cmdName + "'")}}
	}

	err := isValid(validation, cmdName, len(args))

	if hasError(err) {
		return Results{{Error: err}}
	}

	if handler.multEnabled {
		handler.multArgs = append(handler.multArgs, args)
		return Results{{Response: domain.QUEUED}}
	}

	return Results{handler.commands[cmdName](args)}
}
