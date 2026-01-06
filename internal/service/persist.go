package service

import "github.com/luiz-simples/keyp.git/internal/domain"

func (handler *Handler) persist(args Args) *Result {
	res := domain.NewResult()
	key := args[domain.FirstArg]
	removed := handler.storage.Persist(handler.context, key)

	res.Response = formatBool(removed)
	return res
}
