package service

import (
	"context"
	"fmt"
	"net/http"
)

func StartService(ctx context.Context, hostPort string, handleFunc func()) (context.Context, error) {

	handleFunc()

	ctx, cancel := context.WithCancel(ctx)

	var server http.Server
	server.Addr = hostPort

	err := server.ListenAndServe()
	if err != nil {
		fmt.Printf("server listen and serve get err=%v\n", err)
	}
	cancel()

	return ctx, nil
}