package service

import (
	"encoding/binary"

	"github.com/luiz-simples/keyp.git/internal/domain"
)

func (handler *Handler) smembers(args Args) *Result {
	res := domain.NewResult()
	key := args[firstArg]

	members, err := handler.storage.SMembers(handler.context, key)
	if hasError(err) {
		res.Error = err
		return res
	}

	if len(members) == 0 {
		res.Response = make([]byte, 8)
		binary.LittleEndian.PutUint64(res.Response, 0)
		return res
	}

	response := make([]byte, 8)
	binary.LittleEndian.PutUint64(response, uint64(len(members)))

	for _, member := range members {
		response = append(response, make([]byte, 4)...)
		binary.LittleEndian.PutUint32(response[len(response)-4:], uint32(len(member)))
		response = append(response, member...)
	}

	res.Response = response
	return res
}
