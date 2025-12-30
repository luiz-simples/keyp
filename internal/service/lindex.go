package service

import (
	"strconv"

	"github.com/luiz-simples/keyp.git/internal/domain"
)

func (handler *Handler) lindex(args Args) *Result {
	res := domain.NewResult()
	key := args[firstArg]
	indexInt, _ := strconv.ParseInt(string(args[secondArg]), 10, 64)

	value, err := handler.storage.LIndex(handler.context, key, indexInt)
	if hasError(err) {
		res.Error = err
		return res
	}

	res.Response = value
	return res
}
