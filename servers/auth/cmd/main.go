package main

import (
	"auth/internal/handler/health"
	"log"
	"net"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", health.Health)

	listener, err := net.Listen("tcp", ":8080")

	if err != nil {
		log.Fatal(err)
	}

	log.Println("Server running in http://localhost:8080")

	err = http.Serve(listener, mux)

	if err != nil {
		log.Fatal(err)
	}

}
