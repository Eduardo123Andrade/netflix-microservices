package main

import (
	"auth/internal/config"
	"auth/internal/database"
	usersv1 "auth/internal/gen/users/v1"
	"auth/internal/grpcclient"
	"auth/internal/router"
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"
)

func main() {
	c, err := config.Load()

	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	pool, err := database.NewPool(ctx, c.DB)

	if err != nil {
		log.Fatal(err)
	}

	defer pool.Close()

	if err := database.Connect(ctx, pool); err != nil {
		log.Printf("aviso: banco indisponível na subida, seguindo mesmo assim: %v", err)
	}

	userConn, err := grpcclient.NewGrpcUserClient(c.UserServerAddr)

	if err != nil {
		log.Fatal(err)
	}

	defer userConn.Close()

	r := router.New(router.Deps{DB: pool, Cost: c.Cost, Users: usersv1.NewUserServiceClient(userConn)})

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
