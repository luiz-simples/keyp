package service

import (
	"strconv"

	"github.com/luiz-simples/keyp.git/internal/domain"
)

func (handler *Handler) expire(args Args) *Result {
	key := args[firstArg]
	secsInt, _ := strconv.ParseInt(string(args[secondArg]), 10, 64)
	handler.storage.Expire(handler.context, key, uint32(secsInt))
	return domain.NewResult().SetOK()
}
