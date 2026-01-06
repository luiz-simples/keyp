package service

import (
	"github.com/luiz-simples/keyp.git/internal/domain"
)

func (handler *Handler) rpush(args Args) *Result {
	res := domain.NewResult()
	key := args[domain.FirstArg]
	values := args[domain.SecondArg:]

	length := handler.storage.RPush(handler.context, key, values...)

	if length == 0 {
		data, err := handler.storage.Get(handler.context, key)
		if noError(err) && len(data) > 0 {
			res.Error = domain.ErrWrongType
			return res
		}
	}

	res.Response = formatInt64(length)
	return res
}
