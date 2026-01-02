package service

import (
	"encoding/binary"
	"strconv"

	"github.com/luiz-simples/keyp.git/internal/domain"
)

func (handler *Handler) lrange(args Args) *Result {
	res := domain.NewResult()
	key := args[domain.FirstArg]
	startInt, _ := strconv.ParseInt(string(args[domain.SecondArg]), 10, 64)
	stopInt, _ := strconv.ParseInt(string(args[domain.ThirdArg]), 10, 64)

	values, err := handler.storage.LRange(handler.context, key, startInt, stopInt)
	if hasError(err) {
		res.Error = err
		return res
	}

	totalSize := 4
	for _, value := range values {
		totalSize += 4 + len(value)
	}

	res.Response = make([]byte, totalSize)
	offset := 0

	binary.LittleEndian.PutUint32(res.Response[offset:], uint32(len(values)))
	offset += 4

	for _, value := range values {
		binary.LittleEndian.PutUint32(res.Response[offset:], uint32(len(value)))
		offset += 4
		copy(res.Response[offset:], value)
		offset += len(value)
	}

	return res
}
