include .env
export

COMPOSE_DEV := container/docker-compose.dev.yml
ENV := --env-file .env

APP_VERSION := $(shell cat VERSION)
LDFLAGS := -X auth/internal/buildinfo.Version=$(APP_VERSION) 

.PHONY: db-up-dev db-down-dev run

db-up-dev:
	docker compose $(ENV) -f $(COMPOSE_DEV) up -d

db-down-dev:
	docker compose $(ENV) -f $(COMPOSE_DEV) down

run:
	go run -ldflags "$(LDFLAGS)" ./cmd

