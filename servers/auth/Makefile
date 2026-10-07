include .env
export

COMPOSE_DEV := container/docker-compose.dev.yml
ENV := --env-file .env

APP_VERSION := $(shell cat VERSION)
LDFLAGS := -X auth/internal/buildinfo.Version=$(APP_VERSION) 
COVER_PKGS = $(shell go list ./... | grep -v cmd | grep -v buildinfo | paste -sd, -)
MIGRATE_PATH = db/migrations
DB_URL = postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(SSL_MODE)

.PHONY: \
	dev-db-up  \
	dev-db-down  \
	migrate-create \
	dev-migrate-up \
	dev-migrate-down \
	run \
	test \
	test-integration  \
	test-coverage

dev-db-up:
	docker compose $(ENV) -f $(COMPOSE_DEV) up -d --wait
	$(MAKE) dev-migrate-up

dev-db-down:
	docker compose $(ENV) -f $(COMPOSE_DEV) down 

run:
	go run -ldflags "$(LDFLAGS)" ./cmd

test:
	go test ./...

dev-migrate-up:
	migrate -database $(DB_URL) -path $(MIGRATE_PATH) up

dev-migrate-down:
	migrate -database $(DB_URL) -path $(MIGRATE_PATH) down 1

migrate-create:
	$(if $(name),,$(error uso: make dev-migrate-create name=<nome>))
	@migrate create -ext sql -dir $(MIGRATE_PATH) $(name)

test-integration:
	go test -tags integration ./...

test-coverage:
	go test -tags integration -coverpkg $(COVER_PKGS) -coverprofile=cover.out ./...
	go tool cover -func=cover.out
