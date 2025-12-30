package service

import "github.com/luiz-simples/keyp.git/internal/domain"

func (handler *Handler) persist(args Args) *Result {
	key := args[firstArg]
	handler.storage.Persist(handler.context, key)
	return domain.NewResult().SetOK()
}
