package main

import (
	"context"
	"log"

	"github.com/danilovkiri/dk_go_url_shortener/api/rest"
)

func main() {
	ctx := context.Background()
	server, err := rest.InitServer(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(server.ListenAndServe())
}
