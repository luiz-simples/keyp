package service

import (
	"context"
	"sync"

	"github.com/luiz-simples/keyp.git/internal/domain"
)

type (
	Pool struct {
		refs *sync.Pool
	}
)

func NewPool(storage domain.Persister) *Pool {
	return &Pool{
		refs: &sync.Pool{
			New: func() any {
				return NewHandler(storage)
			},
		},
	}
}

func (pool *Pool) Get(ctx context.Context) domain.Dispatcher {
	handler, _ := pool.refs.Get().(*Handler)
	handler.context = context.WithValue(ctx, domain.DB, uint8(0))
	return handler
}

func (pool *Pool) Free(handler domain.Dispatcher) {
	handler.Clear()
	pool.refs.Put(handler)
}
