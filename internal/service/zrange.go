package service

import (
	"strconv"

	"github.com/luiz-simples/keyp.git/internal/domain"
)

func (handler *Handler) zrange(args Args) *Result {
	res := domain.NewResult()
	key := args[domain.FirstArg]
	startStr := string(args[domain.SecondArg])
	stopStr := string(args[domain.ThirdArg])

	start, err := strconv.ParseInt(startStr, 10, 64)
	if hasError(err) {
		res.Error = domain.ErrInvalidInteger
		return res
	}

	stop, err := strconv.ParseInt(stopStr, 10, 64)
	if hasError(err) {
		res.Error = domain.ErrInvalidInteger
		return res
	}

	members, err := handler.storage.ZRange(handler.context, key, start, stop)
	if hasError(err) {
		res.Error = err
		return res
	}

	res.Response = encodeArray(members)
	return res
}
