package main

import (
	"auth/internal/router"
	"log"
	"net"
	"net/http"
)

func main() {
	r := router.New()

	listener, err := net.Listen("tcp", ":8080")

	if err != nil {
		log.Fatal(err)
	}

	log.Println("Server running in http://localhost:8080/api")

	err = http.Serve(listener, r)

	if err != nil {
		log.Fatal(err)
	}

}
