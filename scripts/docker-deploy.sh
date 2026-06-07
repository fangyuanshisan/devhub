#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
COMPOSE_FILE="${COMPOSE_FILE:-docker-compose.deploy.yml}"
ENV_FILE="${ENV_FILE:-.env.docker}"
ACTION="${1:-up}"

cd "$ROOT_DIR"

usage() {
  cat <<'USAGE'
DevHub Docker one-click deployment

Usage:
  ./scripts/docker-deploy.sh up        Build and start DevHub + MySQL
  ./scripts/docker-deploy.sh restart   Rebuild and restart
  ./scripts/docker-deploy.sh status    Show container status
  ./scripts/docker-deploy.sh logs      Follow DevHub logs
  ./scripts/docker-deploy.sh down      Stop containers

Before first run:
  cp .env.docker.example .env.docker
  edit .env.docker and replace every change_me value
USAGE
}

require_env_file() {
  if [[ ! -f "$ENV_FILE" ]]; then
    echo "Missing $ENV_FILE" >&2
    echo "Create it first:" >&2
    echo "  cp .env.docker.example $ENV_FILE" >&2
    echo "  ${EDITOR:-vi} $ENV_FILE" >&2
    exit 2
  fi
  if grep -q 'change_me' "$ENV_FILE"; then
    echo "$ENV_FILE still contains change_me placeholders. Please edit it before deploying." >&2
    exit 2
  fi
}

compose() {
  docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" "$@"
}

case "$ACTION" in
  up)
    require_env_file
    compose up -d --build
    compose ps
    ;;
  restart)
    require_env_file
    compose up -d --build --force-recreate
    compose ps
    ;;
  status)
    require_env_file
    compose ps
    ;;
  logs)
    require_env_file
    compose logs -f devhub
    ;;
  down)
    require_env_file
    compose down
    ;;
  help|-h|--help)
    usage
    ;;
  *)
    echo "Unknown action: $ACTION" >&2
    usage >&2
    exit 2
    ;;
esac
