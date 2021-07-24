package main

import (
	"context"
	"log"
	"net/http"

	"mini_distributed_system/registry"
)

func main() {
	http.Handle("/services", &registry.RegistryService{})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var srv http.Server
	srv.Addr = registry.ServerPort

	go func() {
		log.Println(srv.ListenAndServe())
		cancel()
	}()

	<-ctx.Done()
	log.Println("Shutting down registry service")
}