package service

import "github.com/luiz-simples/keyp.git/internal/domain"

func ping(args Args) *Result {
	result := domain.NewResult()

	if len(args) > 1 {
		result.Response = args[1]
		return result
	}

	result.Response = domain.PONG
	return result
}
