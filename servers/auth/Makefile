COMPOSE_DEV := container/docker-compose.dev.yml
ENV := --env-file .env

.PHONY: db-up-dev db-down-dev run

db-up-dev:
	docker compose $(ENV) -f $(COMPOSE_DEV) up -d

db-down-dev:
	docker compose $(ENV) -f $(COMPOSE_DEV) down

run:
	go run ./cmd

