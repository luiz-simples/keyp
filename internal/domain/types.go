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

		Exists(context.Context, []byte) bool
		LLen(context.Context, []byte) int64
		LIndex(context.Context, []byte, int64) ([]byte, error)
		LSet(context.Context, []byte, int64, []byte) error
		LPush(context.Context, []byte, ...[]byte) int64
		RPush(context.Context, []byte, ...[]byte) int64
		LPop(context.Context, []byte) ([]byte, error)
		RPop(context.Context, []byte) ([]byte, error)
		LRange(context.Context, []byte, int64, int64) ([][]byte, error)

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
