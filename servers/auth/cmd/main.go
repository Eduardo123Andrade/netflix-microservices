package main

import (
	"auth/internal/config"
	"auth/internal/database"
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

	r := router.New(router.Deps{DB: pool})

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
