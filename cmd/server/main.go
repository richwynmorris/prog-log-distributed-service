package main

import (
	"log"

	"github.com/richwynmorris/proglog/internal/server"
)

func main() {
	srv := server.NewHtttpServer(":8080")
	err := srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
