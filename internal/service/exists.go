package service

import (
	"encoding/binary"

	"github.com/luiz-simples/keyp.git/internal/domain"
)

func (handler *Handler) exists(args Args) *Result {
	res := domain.NewResult()
	key := args[firstArg]
	exists := handler.storage.Exists(handler.context, key)

	res.Response = make([]byte, 4)
	value := uint32(0)
	if exists {
		value = 1
	}
	binary.LittleEndian.PutUint32(res.Response, value)

	return res
}
