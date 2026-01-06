package service

import (
	"github.com/luiz-simples/keyp.git/internal/domain"
)

func (handler *Handler) srem(args Args) *Result {
	res := domain.NewResult()
	key := args[domain.FirstArg]
	members := args[domain.SecondArg:]

	count := handler.storage.SRem(handler.context, key, members...)
	res.Response = formatInt64(count)
	return res
}
