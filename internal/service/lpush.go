package service

import (
	"encoding/binary"

	"github.com/luiz-simples/keyp.git/internal/domain"
)

func (handler *Handler) lpush(args Args) *Result {
	res := domain.NewResult()
	key := args[domain.FirstArg]
	values := args[domain.SecondArg:]

	length := handler.storage.LPush(handler.context, key, values...)

	res.Response = make([]byte, 8)
	binary.LittleEndian.PutUint64(res.Response, uint64(length))

	return res
}
