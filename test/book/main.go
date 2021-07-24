package main

import (
	"context"

	"mini_distributed_system/book"
	"mini_distributed_system/service"
)

func main() {
	hostPort := "localhost:8010"

	ctx, err := service.StartService(context.Background(), hostPort, book.InitBookHandler)
	if err != nil {
		panic(err)
	}

	<-ctx.Done()
}
