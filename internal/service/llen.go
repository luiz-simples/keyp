package service

import (
	"github.com/luiz-simples/keyp.git/internal/domain"
)

func (handler *Handler) llen(args Args) *Result {
	res := domain.NewResult()
	key := args[domain.FirstArg]
	length := handler.storage.LLen(handler.context, key)

	res.Response = formatInt64(length)
	return res
}
