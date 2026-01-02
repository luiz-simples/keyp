package main

import (
	"log"

	"github.com/luiz-simples/keyp.git/internal/app"
	"github.com/luiz-simples/keyp.git/internal/service"
	"github.com/luiz-simples/keyp.git/internal/storage"
)

func main() {
	config := app.Config{
		Address: "0.0.0.0:6379",
		DataDir: "./data",
	}

	lmdb, err := storage.NewClient(config.DataDir)

	if noError(err) {
		poolService := service.NewPool(lmdb)
		server := app.NewServer(poolService)
		defer server.Close()
		err = server.Start(config)
	}

	if err != nil {
		log.Fatal(err)
	}
}

func noError(err error) bool {
	return err == nil
}
