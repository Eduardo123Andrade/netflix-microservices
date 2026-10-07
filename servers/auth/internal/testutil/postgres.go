//go:build integration

// Package testutil reúne utilitários usados só pelos testes de integração.
package testutil

import (
	"auth/internal/config"
	"context"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

// Mesma imagem do container/docker-compose.dev.yml.
const postgresImage = "postgres:18.0-alpine3.22"

// StartPostgres sobe um Postgres descartável e devolve o container e a
// config para conectar nele. O container é destruído no fim do teste.
func StartPostgres(t *testing.T) (*postgres.PostgresContainer, config.Database) {
	t.Helper()
	ctx := context.Background()

	cfg := config.Database{
		User:     "test_user",
		Password: "test_password",
		Name:     "test_db",
	}

	ctr, err := postgres.Run(ctx, postgresImage,
		postgres.WithDatabase(cfg.Name),
		postgres.WithUsername(cfg.User),
		postgres.WithPassword(cfg.Password),
		postgres.BasicWaitStrategies(),
	)
	testcontainers.CleanupContainer(t, ctr)
	if err != nil {
		t.Fatalf("start postgres: %v", err)
	}

	host, err := ctr.Host(ctx)
	if err != nil {
		t.Fatalf("container host: %v", err)
	}
	port, err := ctr.MappedPort(ctx, "5432/tcp")
	if err != nil {
		t.Fatalf("container port: %v", err)
	}

	cfg.Host = host
	cfg.Port = int(port.Num())

	return ctr, cfg
}
