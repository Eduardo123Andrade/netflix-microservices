package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"auth/internal/config"
	"auth/internal/database"
	"auth/internal/grpcclient"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
)

// startupTimeout limita quanto a subida espera pelo banco.
const startupTimeout = 5 * time.Second

// connections reúne o que o serviço abre para fora na subida.
type connections struct {
	db    *pgxpool.Pool
	users *grpc.ClientConn
}

// openConnections abre o banco e o cliente gRPC do users. O close devolvido
// fecha tudo na ordem inversa da abertura; o main chama com defer.
func openConnections(c config.Config) (connections, func(), error) {
	ctx, cancel := context.WithTimeout(context.Background(), startupTimeout)
	defer cancel()

	pool, err := database.NewPool(ctx, c.DB)
	if err != nil {
		return connections{}, nil, fmt.Errorf("open database: %w", err)
	}

	if err := database.Connect(ctx, pool); err != nil {
		log.Printf("aviso: banco indisponível na subida, seguindo mesmo assim: %v", err)
	}

	users, err := grpcclient.NewGrpcUserClient(c.UserServerAddr)
	if err != nil {
		// O pool já foi aberto; fecha antes de desistir.
		pool.Close()
		return connections{}, nil, fmt.Errorf("open users client: %w", err)
	}

	closeAll := func() {
		if err := users.Close(); err != nil {
			log.Printf("aviso: erro ao fechar o cliente do users: %v", err)
		}
		pool.Close()
	}

	return connections{db: pool, users: users}, closeAll, nil
}
