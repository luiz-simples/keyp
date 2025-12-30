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

		"EXISTS": handler.exists,
		"LLEN":   handler.llen,
		"LINDEX": handler.lindex,
		"LSET":   handler.lset,
		"LPUSH":  handler.lpush,
		"RPUSH":  handler.rpush,
		"LPOP":   handler.lpop,
		"RPOP":   handler.rpop,
		"LRANGE": handler.lrange,

		"FLUSHALL":  handler.flushall,
		"SADD":      handler.sadd,
		"SREM":      handler.srem,
		"SMEMBERS":  handler.smembers,
		"SISMEMBER": handler.sismember,

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

		"EXISTS": {MinArgs: 2, MaxArgs: 2},
		"LLEN":   {MinArgs: 2, MaxArgs: 2},
		"LINDEX": {MinArgs: 3, MaxArgs: 3},
		"LSET":   {MinArgs: 4, MaxArgs: 4},
		"LPUSH":  {MinArgs: 3, MaxArgs: -1},
		"RPUSH":  {MinArgs: 3, MaxArgs: -1},
		"LPOP":   {MinArgs: 2, MaxArgs: 2},
		"RPOP":   {MinArgs: 2, MaxArgs: 2},
		"LRANGE": {MinArgs: 4, MaxArgs: 4},

		"FLUSHALL":  {MinArgs: 1, MaxArgs: 1},
		"SADD":      {MinArgs: 3, MaxArgs: -1},
		"SREM":      {MinArgs: 3, MaxArgs: -1},
		"SMEMBERS":  {MinArgs: 2, MaxArgs: 2},
		"SISMEMBER": {MinArgs: 3, MaxArgs: 3},

		"PING":   {MinArgs: 1, MaxArgs: 2},
		"DELETE": {MinArgs: 2, MaxArgs: -1},
	}

	return handler
}
