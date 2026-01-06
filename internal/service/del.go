package service

import (
	"github.com/luiz-simples/keyp.git/internal/domain"
)

func (handler *Handler) del(args Args) *Result {
	if len(args) < 2 {
		res := domain.NewResult()
		res.Error = newInvalidArgsError("DEL")
		return res
	}

	keys := args[domain.FirstArg:]
	deleted, err := handler.storage.Del(handler.context, keys...)
	res := domain.NewResult()

	if isContextCanceled(err) {
		return res.SetCanceled()
	}

	res.Response = formatUint32(deleted)
	return res
}
