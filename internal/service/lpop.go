package service

import "github.com/luiz-simples/keyp.git/internal/domain"

func (handler *Handler) lpop(args Args) *Result {
	res := domain.NewResult()
	key := args[firstArg]

	value, err := handler.storage.LPop(handler.context, key)
	if hasError(err) {
		res.Error = err
		return res
	}

	res.Response = value
	return res
}
