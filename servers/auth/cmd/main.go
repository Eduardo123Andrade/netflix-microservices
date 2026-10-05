package main

import (
	"auth/internal/config"
	"auth/internal/router"
	"fmt"
	"log"
	"net"
	"net/http"
)

func main() {
	c, err := config.Load()

	if err != nil {
		log.Fatal(err)
	}

	r := router.New()

	p := fmt.Sprintf(":%d", c.Port)
	listener, err := net.Listen("tcp", p)

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Server running in http://localhost:%d/api\n", c.Port)

	err = http.Serve(listener, r)

	if err != nil {
		log.Fatal(err)
	}

}
