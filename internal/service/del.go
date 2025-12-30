package service

import (
	"encoding/binary"

	"github.com/luiz-simples/keyp.git/internal/domain"
)

func (handler *Handler) del(args Args) *Result {
	keys := args[firstArg:]
	deleted, err := handler.storage.Del(handler.context, keys...)
	res := domain.NewResult()

	if isContextCanceled(err) {
		return res.SetCanceled()
	}

	res.Response = make([]byte, 4)
	binary.LittleEndian.PutUint32(res.Response, deleted)

	return res
}
