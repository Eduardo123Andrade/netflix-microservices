#!/usr/bin/env bash
set -euo pipefail

# Derruba o Postgres de um ambiente com `docker compose down` (SIGTERM e
# timeout padrão, nunca kill).
#
# Uso:
#   bash scripts/db-down.sh dev
#   bash scripts/db-down.sh integration
#   bash scripts/db-down.sh e2e
#   bash scripts/db-down.sh all

cd "$(dirname "$0")/.."

ENV="${1:-}"

down_env() {
  local env_name="$1"
  echo "→ Derrubando Postgres '$env_name'..."

  # integration/e2e são efêmeros: -v garante que nenhum volume sobrevive à
  # execução. dev mantém os dados entre sessões, então não leva -v.
  # O dev interpola variáveis do .env no compose; sem ele o `down` falharia
  # na validação, por isso o fallback para o .env.example.
  case "$env_name" in
    dev)
      local env_file=".env"
      [ -f "$env_file" ] || env_file=".env.example"
      docker compose --env-file "$env_file" -f docker-compose.dev.yml down
      ;;
    integration|e2e)
      docker compose -f "docker-compose.$env_name.yml" down -v
      ;;
  esac

  echo "✓ Postgres '$env_name' encerrado."
}

case "$ENV" in
  dev|integration|e2e)
    down_env "$ENV"
    ;;
  all)
    for e in dev integration e2e; do
      down_env "$e"
    done
    ;;
  *)
    echo "Uso: bash scripts/db-down.sh <dev|integration|e2e|all>"
    exit 1
    ;;
esac
