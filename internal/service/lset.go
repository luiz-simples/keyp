package service

import (
	"strconv"

	"github.com/luiz-simples/keyp.git/internal/domain"
)

func (handler *Handler) lset(args Args) *Result {
	res := domain.NewResult()
	key := args[domain.FirstArg]
	indexInt, _ := strconv.ParseInt(string(args[domain.SecondArg]), 10, 64)
	value := args[domain.ThirdArg]

	err := handler.storage.LSet(handler.context, key, indexInt, value)
	if hasError(err) {
		res.Error = err
		return res
	}

	res.Response = OK
	return res
}
