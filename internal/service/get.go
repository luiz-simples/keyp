package service

import "github.com/luiz-simples/keyp.git/internal/domain"

func (handler *Handler) get(args Args) *Result {
	key := args[domain.FirstArg]
	res := domain.NewResult()
	res.Response, res.Error = handler.storage.Get(handler.context, key)

	if isContextCanceled(res.Error) {
		return res.SetCanceled()
	}

	if isKeyNotFoundError(res.Error) {
		return res.Clear()
	}

	return res
}
