package service

import (
	"encoding/binary"

	"github.com/luiz-simples/keyp.git/internal/domain"
)

func (handler *Handler) incr(args Args) *Result {
	res := domain.NewResult()
	key := args[domain.FirstArg]

	value, err := handler.storage.Incr(handler.context, key)
	if hasError(err) {
		res.Error = err
		return res
	}

	res.Response = make([]byte, 8)
	binary.LittleEndian.PutUint64(res.Response, uint64(value))

	return res
}
