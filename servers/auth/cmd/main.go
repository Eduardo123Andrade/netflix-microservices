package main

import (
	"fmt"
	"log"
	"net"
	"net/http"

	"auth/internal/config"
	usersv1 "auth/internal/gen/users/v1"
	"auth/internal/router"
)

func main() {
	c, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	// Sem chave o auth não cumpre a função principal: falha antes de abrir
	// qualquer conexão.
	tokens, err := newTokenService(c.JWT)
	if err != nil {
		log.Fatal(err)
	}

	conns, closeConns, err := openConnections(c)
	if err != nil {
		log.Fatal(err)
	}
	defer closeConns()

	r, err := router.New(router.Deps{
		DB:         conns.db,
		Cost:       c.Cost,
		Users:      usersv1.NewUserServiceClient(conns.users),
		Tokens:     tokens,
		RefreshTTL: c.JWT.RefreshTTL,
	})
	if err != nil {
		log.Fatal(err)
	}

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
