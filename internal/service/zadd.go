package service

import (
	"encoding/binary"
	"strconv"

	"github.com/luiz-simples/keyp.git/internal/domain"
)

func (handler *Handler) zadd(args Args) *Result {
	res := domain.NewResult()
	key := args[firstArg]
	scoreStr := string(args[secondArg])
	member := args[thirdArg]

	score, err := strconv.ParseFloat(scoreStr, 64)
	if hasError(err) {
		res.Error = ErrInvalidFloat
		return res
	}

	count := handler.storage.ZAdd(handler.context, key, score, member)

	res.Response = make([]byte, 8)
	binary.LittleEndian.PutUint64(res.Response, uint64(count))

	return res
}
