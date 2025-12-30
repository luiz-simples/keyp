package service

import "github.com/luiz-simples/keyp.git/internal/domain"

func (handler *Handler) flushall(args Args) *Result {
	res := domain.NewResult()

	err := handler.storage.FlushAll(handler.context)
	if hasError(err) {
		res.Error = err
		return res
	}

	res.Response = []byte("OK")
	return res
}
