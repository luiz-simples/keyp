package service

import (
	"github.com/luiz-simples/keyp.git/internal/domain"
)

func (handler *Handler) sadd(args Args) *Result {
	res := domain.NewResult()
	key := args[domain.FirstArg]
	members := args[domain.SecondArg:]

	count := handler.storage.SAdd(handler.context, key, members...)
	res.Response = formatInt64(count)
	return res
}
