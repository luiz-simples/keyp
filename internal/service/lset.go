package service

import (
	"strconv"

	"github.com/luiz-simples/keyp.git/internal/domain"
)

func (handler *Handler) lset(args Args) *Result {
	res := domain.NewResult()
	key := args[firstArg]
	indexInt, _ := strconv.ParseInt(string(args[secondArg]), 10, 64)
	value := args[thirdArg]

	err := handler.storage.LSet(handler.context, key, indexInt, value)
	if hasError(err) {
		res.Error = err
		return res
	}

	res.Response = OK
	return res
}
