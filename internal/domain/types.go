package domain

import (
	"context"
	"errors"
)

var (
	OK     []byte = []byte("OK")
	PONG   []byte = []byte("PONG")
	QUEUED []byte = []byte("QUEUED")

	ErrEmpty          error = errors.New("ERR empty command")
	ErrCanceled       error = errors.New("ERR operation canceled")
	ErrInvalidFloat   error = errors.New("ERR value is not a valid float")
	ErrInvalidInteger error = errors.New("ERR value is not an integer or out of range")
	ErrWrongType      error = errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
)

const (
	PING    string = "PING"
	MULTI   string = "MULTI"
	EXEC    string = "EXEC"
	DISCARD string = "DISCARD"

	EmptyArgs  = 0
	CommandArg = 0
	FirstArg   = 1
	SecondArg  = 2
	ThirdArg   = 3
)

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
		Persist(context.Context, []byte) bool
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

		FlushAll(context.Context) error
		SAdd(context.Context, []byte, ...[]byte) int64
		SRem(context.Context, []byte, ...[]byte) int64
		SMembers(context.Context, []byte) ([][]byte, error)
		SIsMember(context.Context, []byte, []byte) bool

		ZAdd(context.Context, []byte, float64, []byte) int64
		ZRange(context.Context, []byte, int64, int64) ([][]byte, error)
		ZCount(context.Context, []byte, float64, float64) int64

		Incr(context.Context, []byte) (int64, error)
		IncrBy(context.Context, []byte, int64) (int64, error)
		Decr(context.Context, []byte) (int64, error)
		DecrBy(context.Context, []byte, int64) (int64, error)

		Append(context.Context, []byte, []byte) int64

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
