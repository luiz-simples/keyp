package service

import "github.com/luiz-simples/keyp.git/internal/domain"

func (handler *Handler) rpop(args Args) *Result {
	res := domain.NewResult()
	key := args[domain.FirstArg]

	value, err := handler.storage.RPop(handler.context, key)
	if hasError(err) {
		res.Error = err
		return res
	}

	res.Response = value
	return res
}
