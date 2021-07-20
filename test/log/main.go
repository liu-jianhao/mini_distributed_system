package main

import (
	"context"

	"mini_distributed_system/log"
	"mini_distributed_system/service"
)

func main() {
	log.InitLogger("test.log")

	hostPort := "localhost:8000"

	ctx, err := service.StartService(context.Background(), hostPort, log.RegisterLogHandler)
	if err != nil {
		panic(err)
	}

	<-ctx.Done()
}
