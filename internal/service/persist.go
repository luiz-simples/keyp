package service

import "github.com/luiz-simples/keyp.git/internal/domain"

func (handler *Handler) persist(args Args) *Result {
	key := args[domain.FirstArg]
	handler.storage.Persist(handler.context, key)
	return domain.NewResult().SetOK()
}
