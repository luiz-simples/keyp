package service

import (
	"encoding/binary"
	"strconv"

	"github.com/luiz-simples/keyp.git/internal/domain"
)

func (handler *Handler) decrby(args Args) *Result {
	res := domain.NewResult()
	key := args[firstArg]
	decrementStr := string(args[secondArg])

	decrement, err := strconv.ParseInt(decrementStr, 10, 64)
	if hasError(err) {
		res.Error = ErrInvalidInteger
		return res
	}

	value, err := handler.storage.DecrBy(handler.context, key, decrement)
	if hasError(err) {
		res.Error = err
		return res
	}

	res.Response = make([]byte, 8)
	binary.LittleEndian.PutUint64(res.Response, uint64(value))

	return res
}
