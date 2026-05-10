#!/usr/bin/env bash
set -euo pipefail

if [ -d /workspace/web/frontend-app ]; then
  cd /workspace/web/frontend-app
fi

if [ ! -d node_modules ] || [ ! -x node_modules/.bin/playwright ]; then
  echo "frontend-e2e node_modules missing in mounted workspace; restoring dependencies from image layer"
  mkdir -p ./node_modules
  cp -a /app/node_modules/. ./node_modules/
fi

exec "$@"
