package service

import (
	"encoding/binary"

	"github.com/luiz-simples/keyp.git/internal/domain"
)

func (handler *Handler) llen(args Args) *Result {
	res := domain.NewResult()
	key := args[domain.FirstArg]
	length := handler.storage.LLen(handler.context, key)

	res.Response = make([]byte, 8)
	binary.LittleEndian.PutUint64(res.Response, uint64(length))

	return res
}
