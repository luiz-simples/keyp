package service

import (
	"strconv"

	"github.com/luiz-simples/keyp.git/internal/domain"
)

func (handler *Handler) zadd(args Args) *Result {
	res := domain.NewResult()
	key := args[domain.FirstArg]
	scoreStr := string(args[domain.SecondArg])
	member := args[domain.ThirdArg]

	score, err := strconv.ParseFloat(scoreStr, 64)
	if hasError(err) {
		res.Error = domain.ErrInvalidFloat
		return res
	}

	count := handler.storage.ZAdd(handler.context, key, score, member)
	res.Response = formatInt64(count)
	return res
}
