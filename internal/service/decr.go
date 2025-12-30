package service

import (
	"encoding/binary"

	"github.com/luiz-simples/keyp.git/internal/domain"
)

func (handler *Handler) decr(args Args) *Result {
	res := domain.NewResult()
	key := args[firstArg]

	value, err := handler.storage.Decr(handler.context, key)
	if hasError(err) {
		res.Error = err
		return res
	}

	res.Response = make([]byte, 8)
	binary.LittleEndian.PutUint64(res.Response, uint64(value))

	return res
}
