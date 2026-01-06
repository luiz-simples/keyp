package service

import (
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

	res.Response = formatArray(values)
	return res
}
