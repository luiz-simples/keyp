package service

import (
	"encoding/binary"
	"strconv"

	"github.com/luiz-simples/keyp.git/internal/domain"
)

func (handler *Handler) zcount(args Args) *Result {
	res := domain.NewResult()
	key := args[firstArg]
	minStr := string(args[secondArg])
	maxStr := string(args[thirdArg])

	min, err := strconv.ParseFloat(minStr, 64)
	if hasError(err) {
		res.Error = ErrInvalidFloat
		return res
	}

	max, err := strconv.ParseFloat(maxStr, 64)
	if hasError(err) {
		res.Error = ErrInvalidFloat
		return res
	}

	count := handler.storage.ZCount(handler.context, key, min, max)

	res.Response = make([]byte, 8)
	binary.LittleEndian.PutUint64(res.Response, uint64(count))

	return res
}
