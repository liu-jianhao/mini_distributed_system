package main

import (
	"context"

	"mini_distributed_system/book"
	"mini_distributed_system/registry"
	"mini_distributed_system/service"
)

func main() {
	hostPort := "localhost:8010"

	reg := &registry.Registration{
		ServiceName:      registry.BookService,
		ServiceURL:       "http://" + hostPort,
		RequiredServices: []registry.ServiceName{registry.LogService},
		ServiceUpdateURL: "http://" + hostPort + "/services",
	}

	ctx, err := service.StartService(context.Background(), hostPort, reg, book.InitBookHandler)
	if err != nil {
		panic(err)
	}

	<-ctx.Done()
}
