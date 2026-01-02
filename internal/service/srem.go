package service

import (
	"encoding/binary"

	"github.com/luiz-simples/keyp.git/internal/domain"
)

func (handler *Handler) srem(args Args) *Result {
	res := domain.NewResult()
	key := args[domain.FirstArg]
	members := args[domain.SecondArg:]

	count := handler.storage.SRem(handler.context, key, members...)

	res.Response = make([]byte, 8)
	binary.LittleEndian.PutUint64(res.Response, uint64(count))

	return res
}
