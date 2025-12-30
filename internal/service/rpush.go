package service

import (
	"encoding/binary"

	"github.com/luiz-simples/keyp.git/internal/domain"
)

func (handler *Handler) rpush(args Args) *Result {
	res := domain.NewResult()
	key := args[firstArg]
	values := args[secondArg:]

	length := handler.storage.RPush(handler.context, key, values...)

	res.Response = make([]byte, 8)
	binary.LittleEndian.PutUint64(res.Response, uint64(length))

	return res
}
