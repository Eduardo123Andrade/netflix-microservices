#!/usr/bin/env bash
set -euo pipefail

# Sobe o banco de um ambiente de teste (integration ou e2e), roda o comando
# informado e SEMPRE derruba o banco no final (sucesso, falha ou Ctrl+C) via
# trap em EXIT. Assim nenhum banco "fantasma" fica de pé com dados de uma
# execução anterior.
#
# Uso:
#   bash scripts/run-with-db.sh <integration|e2e> -- <comando...>

cd "$(dirname "$0")/.."

ENV="${1:-}"
shift || true

if [ "${1:-}" = "--" ]; then
  shift
fi

case "$ENV" in
  integration|e2e) ;;
  *)
    echo "Uso: bash scripts/run-with-db.sh <integration|e2e> -- <comando...>"
    exit 1
    ;;
esac

if [ "$#" -eq 0 ]; then
  echo "Uso: bash scripts/run-with-db.sh <integration|e2e> -- <comando...>"
  exit 1
fi

cleanup() {
  bash scripts/db-down.sh "$ENV"
}
trap cleanup EXIT

bash scripts/db-up.sh "$ENV"

"$@"
