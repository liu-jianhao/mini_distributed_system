package main

import (
	"context"

	"mini_distributed_system/log"
	"mini_distributed_system/registry"
	"mini_distributed_system/service"
)

func main() {
	log.InitLogger("test.log")

	hostPort := "localhost:8000"

	reg := &registry.Registration{
		ServiceName:      registry.LogService,
		ServiceURL:       "http://" + hostPort,
		RequiredServices: make([]registry.ServiceName, 0),
		ServiceUpdateURL: "http://" + hostPort + "/services",
	}

	ctx, err := service.StartService(context.Background(), hostPort, reg, log.RegisterLogHandler)
	if err != nil {
		panic(err)
	}

	<-ctx.Done()
}
