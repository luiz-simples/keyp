package domain

import "context"

type (
	Result struct {
		Error    error
		Response []byte
	}

	Command  func(Args) *Result
	Args     = [][]byte
	Results  []*Result
	Commands map[string]Command

	Persister interface {
		Set(context.Context, []byte, []byte) error
		Get(context.Context, []byte) ([]byte, error)
		Del(context.Context, ...[]byte) (uint32, error)

		TTL(context.Context, []byte) uint32
		Persist(context.Context, []byte)
		Expire(context.Context, []byte, uint32)

		Close()
	}

	Dispatcher interface {
		Apply(ctx context.Context, args Args) Results
		Clear()
	}

	Logicaler interface {
		Get(ctx context.Context) Dispatcher
		Free(handler Dispatcher)
	}

	Validation struct {
		MinArgs int
		MaxArgs int
	}

	Validations map[string]*Validation

	CTX string
)

const (
	DB = CTX("DB")
	ID = CTX("ID")
)
