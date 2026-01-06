package service

import (
	"github.com/luiz-simples/keyp.git/internal/domain"
)

func (handler *Handler) smembers(args Args) *Result {
	res := domain.NewResult()
	key := args[domain.FirstArg]

	members, err := handler.storage.SMembers(handler.context, key)
	if hasError(err) {
		res.Error = err
		return res
	}

	res.Response = formatArray(members)
	return res
}
