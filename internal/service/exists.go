package service

import (
	"github.com/luiz-simples/keyp.git/internal/domain"
)

func (handler *Handler) exists(args Args) *Result {
	res := domain.NewResult()
	keys := args[domain.FirstArg:]
	count := int64(0)

	for _, key := range keys {
		if handler.storage.Exists(handler.context, key) {
			count++
		}
	}

	res.Response = formatInt64(count)
	return res
}
