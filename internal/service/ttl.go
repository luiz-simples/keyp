package service

import (
	"github.com/luiz-simples/keyp.git/internal/domain"
)

func (handler *Handler) ttl(args Args) *Result {
	res := domain.NewResult()
	key := args[domain.FirstArg]
	secs := handler.storage.TTL(handler.context, key)

	if secs == 0 {
		if handler.storage.Exists(handler.context, key) {
			res.Response = []byte("-1")
		} else {
			res.Response = []byte("-2")
		}
		return res
	}

	if secs == 0xFFFFFFFF {
		res.Response = []byte("-1")
		return res
	}

	res.Response = formatUint32(secs)
	return res
}
