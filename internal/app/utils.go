package app

import (
	"time"

	"github.com/luiz-simples/keyp.git/internal/domain"
)

func hasError(err error) bool {
	return err != nil
}

func hasResponse(response []byte) bool {
	return response != nil
}

func generateConnectionID() int64 {
	return time.Now().UnixNano()
}

func handlerExists(handlers map[int64]domain.Dispatcher, connID int64) bool {
	_, exists := handlers[connID]
	return exists
}

func isArrayResponse(response []byte) bool {
	return len(response) > 0 && response[0] == '*'
}
