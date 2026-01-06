package service

import (
	"github.com/luiz-simples/keyp.git/internal/domain"
)

func (handler *Handler) decr(args Args) *Result {
	res := domain.NewResult()
	key := args[domain.FirstArg]

	value, err := handler.storage.Decr(handler.context, key)
	if hasError(err) {
		res.Error = err
		return res
	}

	res.Response = formatInt64(value)
	return res
}
