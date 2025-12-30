package service

import (
	"encoding/binary"
	"strconv"

	"github.com/luiz-simples/keyp.git/internal/domain"
)

func (handler *Handler) incrby(args Args) *Result {
	res := domain.NewResult()
	key := args[firstArg]
	incrementStr := string(args[secondArg])

	increment, err := strconv.ParseInt(incrementStr, 10, 64)
	if hasError(err) {
		res.Error = ErrInvalidInteger
		return res
	}

	value, err := handler.storage.IncrBy(handler.context, key, increment)
	if hasError(err) {
		res.Error = err
		return res
	}

	res.Response = make([]byte, 8)
	binary.LittleEndian.PutUint64(res.Response, uint64(value))

	return res
}
