package service

import (
	"github.com/luiz-simples/keyp.git/internal/domain"
)

func (handler *Handler) sismember(args Args) *Result {
	res := domain.NewResult()
	key := args[domain.FirstArg]
	member := args[domain.SecondArg]

	exists := handler.storage.SIsMember(handler.context, key, member)
	res.Response = formatBool(exists)
	return res
}
