package service

import (
	"context"

	"github.com/luiz-simples/keyp.git/internal/domain"
)

var OK = domain.OK

type (
	Args    = domain.Args
	Result  = domain.Result
	Results = domain.Results

	Handler struct {
		context context.Context
		storage domain.Persister

		commands    domain.Commands
		validations domain.Validations

		multArgs    []Args
		multEnabled bool
	}
)

func NewHandler(storage domain.Persister) *Handler {
	ctx := context.WithValue(context.Background(), domain.DB, uint8(0))

	handler := &Handler{
		context:     ctx,
		storage:     storage,
		multArgs:    make([]Args, 0),
		multEnabled: false,
	}

	handler.commands = domain.Commands{
		"DEL": handler.del,
		"GET": handler.get,
		"SET": handler.set,
		"SEL": handler.sel,

		"TTL":     handler.ttl,
		"EXPIRE":  handler.expire,
		"PERSIST": handler.persist,

		"PING":   ping,
		"DELETE": handler.del,
	}

	handler.validations = domain.Validations{
		"DEL": {MinArgs: 2, MaxArgs: -1},
		"GET": {MinArgs: 2, MaxArgs: 2},
		"SET": {MinArgs: 3, MaxArgs: 3},
		"SEL": {MinArgs: 2, MaxArgs: 2},

		"TTL":     {MinArgs: 2, MaxArgs: 2},
		"EXPIRE":  {MinArgs: 3, MaxArgs: 3},
		"PERSIST": {MinArgs: 2, MaxArgs: 2},

		"PING":   {MinArgs: 1, MaxArgs: 2},
		"DELETE": {MinArgs: 2, MaxArgs: -1},
	}

	return handler
}
