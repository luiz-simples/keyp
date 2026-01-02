package service

import (
	"encoding/binary"

	"github.com/luiz-simples/keyp.git/internal/domain"
)

func (handler *Handler) append(args Args) *Result {
	res := domain.NewResult()
	key := args[domain.FirstArg]
	value := args[domain.SecondArg]

	length := handler.storage.Append(handler.context, key, value)

	res.Response = make([]byte, 8)
	binary.LittleEndian.PutUint64(res.Response, uint64(length))

	return res
}
