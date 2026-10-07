include .env
export

COMPOSE_DEV := container/docker-compose.dev.yml
ENV := --env-file .env

APP_VERSION := $(shell cat VERSION)
LDFLAGS := -X auth/internal/buildinfo.Version=$(APP_VERSION) 

.PHONY: dev-db-up dev-db-down run test test-integration

dev-db-up:
	docker compose $(ENV) -f $(COMPOSE_DEV) up -d

dev-db-down:
	docker compose $(ENV) -f $(COMPOSE_DEV) down

run:
	go run -ldflags "$(LDFLAGS)" ./cmd

test:
	go test ./...

test-integration:
	go test -tags integration ./...
