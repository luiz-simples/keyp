package app

import (
	"github.com/luiz-simples/keyp.git/internal/domain"

	"github.com/bwmarrin/snowflake"
)

func hasError(err error) bool {
	return err != nil
}

func hasResponse(response []byte) bool {
	return len(response) > 0
}

func generateConnectionID() int64 {
	node, _ := snowflake.NewNode(1)
	return node.Generate().Int64()
}

func contextExists(contexts map[int64]func(), connID int64) bool {
	_, exists := contexts[connID]
	return exists
}

func handlerExists(handlers map[int64]domain.Dispatcher, connID int64) bool {
	_, exists := handlers[connID]
	return exists
}
