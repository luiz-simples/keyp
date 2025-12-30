package service

import "github.com/luiz-simples/keyp.git/internal/domain"

func (handler *Handler) set(args Args) *Result {
	key := args[firstArg]
	value := args[secondArg]

	res := domain.NewResult()
	res.Error = handler.storage.Set(handler.context, key, value)

	if isContextCanceled(res.Error) {
		return res.SetCanceled()
	}

	if hasError(res.Error) {
		return res
	}

	return res.SetOK()
}
