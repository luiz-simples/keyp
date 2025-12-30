package service

import (
	"encoding/binary"

	"github.com/luiz-simples/keyp.git/internal/domain"
)

func (handler *Handler) sismember(args Args) *Result {
	res := domain.NewResult()
	key := args[firstArg]
	member := args[secondArg]

	exists := handler.storage.SIsMember(handler.context, key, member)

	res.Response = make([]byte, 4)
	value := uint32(0)
	if exists {
		value = 1
	}
	binary.LittleEndian.PutUint32(res.Response, value)

	return res
}
