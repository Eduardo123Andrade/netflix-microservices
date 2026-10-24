#!/usr/bin/env bash
set -euo pipefail

# Sobe o Postgres de um ambiente (dev, integration, e2e), espera o healthcheck
# do container ficar "healthy" (docker compose --wait, sem sleep fixo) e roda
# as migrations do Prisma nele.
#
# Uso:
#   bash scripts/db-up.sh dev
#   bash scripts/db-up.sh integration
#   bash scripts/db-up.sh e2e

cd "$(dirname "$0")/.."

ENV="${1:-}"

case "$ENV" in
  dev)
    if [ ! -f .env ]; then
      echo "✗ servers/users/.env não existe. Copie o .env.example e ajuste."
      exit 1
    fi
    ENV_FILE=".env"
    ;;
  integration|e2e)
    ENV_FILE=".env.$ENV"
    ;;
  *)
    echo "Uso: bash scripts/db-up.sh <dev|integration|e2e>"
    exit 1
    ;;
esac

echo "→ Subindo Postgres '$ENV'..."
docker compose --env-file "$ENV_FILE" -f "docker-compose.$ENV.yml" up -d --wait

echo "  Rodando migrations ($ENV)..."
pnpm exec dotenv -e "$ENV_FILE" -- prisma migrate deploy

echo "✓ Ambiente '$ENV' pronto."
