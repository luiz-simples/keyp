package service

import (
	"strconv"

	"github.com/luiz-simples/keyp.git/internal/domain"
)

func (handler *Handler) set(args Args) *Result {
	key := args[domain.FirstArg]
	value := args[domain.SecondArg]

	res := domain.NewResult()
	res.Error = handler.storage.Set(handler.context, key, value)

	if isContextCanceled(res.Error) {
		return res.SetCanceled()
	}

	if hasError(res.Error) {
		return res
	}

	if len(args) >= 5 {
		option := string(args[3])
		ttlStr := string(args[4])

		if option == "EX" {
			ttl, err := strconv.ParseInt(ttlStr, 10, 64)
			if hasError(err) {
				res.Error = domain.ErrInvalidInteger
				return res
			}
			handler.storage.Expire(handler.context, key, uint32(ttl))
		}

		if option == "PX" {
			ttl, err := strconv.ParseInt(ttlStr, 10, 64)
			if hasError(err) {
				res.Error = domain.ErrInvalidInteger
				return res
			}
			handler.storage.Expire(handler.context, key, uint32(ttl/1000))
		}
	}

	return res.SetOK()
}
