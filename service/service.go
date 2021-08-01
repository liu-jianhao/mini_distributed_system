package service

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"mini_distributed_system/registry"
)

func StartService(ctx context.Context, hostPort string, reg *registry.Registration, handleFunc func()) (context.Context, error) {

	handleFunc()

	ctx, cancel := context.WithCancel(ctx)

	var server http.Server
	var err error

	server.Addr = hostPort

	go func() {
		err = server.ListenAndServe()
		if err != nil {
			log.Printf("server listen and serve get err=%v\n\n", err)
		}
		err = registry.ShutdownService(fmt.Sprintf("http://%s", hostPort))
		if err != nil {
			log.Println(err)
		}
		cancel()
	}()

	go func() {
		fmt.Printf("%v started. Press any key to stop. \n", reg.ServiceName)
		var s string
		_, _ = fmt.Scanln(&s)
		err := registry.ShutdownService(fmt.Sprintf("http://%s", hostPort))
		if err != nil {
			log.Println(err)
		}
		_ = server.Shutdown(ctx)
		cancel()
	}()

	err = registry.RegisterService(reg)
	if err != nil {
		log.Println(err)
		return ctx, err
	}

	return ctx, nil
}
