package service

import (
	"encoding/binary"

	"github.com/luiz-simples/keyp.git/internal/domain"
)

func (handler *Handler) ttl(args Args) *Result {
	res := domain.NewResult()
	key := args[firstArg]
	secs := handler.storage.TTL(handler.context, key)

	res.Response = make([]byte, 4)
	binary.LittleEndian.PutUint32(res.Response, secs)

	return res
}
