package service

import (
	"context"
	"strconv"

	"github.com/luiz-simples/keyp.git/internal/domain"
)

func (handler *Handler) sel(args Args) *Result {
	res := domain.NewResult()
	dbInt, _ := strconv.ParseInt(string(args[domain.FirstArg]), 10, 64)
	handler.context = context.WithValue(handler.context, domain.DB, uint8(dbInt))
	return res.SetOK()
}
