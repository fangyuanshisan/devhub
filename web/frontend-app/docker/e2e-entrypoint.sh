#!/usr/bin/env bash
set -euo pipefail

if [ -d /workspace/web/frontend-app ]; then
  cd /workspace/web/frontend-app
fi

if [ ! -d node_modules ] || [ ! -x node_modules/.bin/playwright ]; then
  echo "frontend-e2e node_modules missing in mounted workspace; restoring dependencies from image layer"
  if ! mkdir -p ./node_modules 2>/dev/null || ! cp -R /app/node_modules/. ./node_modules/ 2>/dev/null; then
    echo "workspace node_modules is not writable; using image layer dependencies"
    export PATH="/app/node_modules/.bin:$PATH"
    export NODE_PATH="/app/node_modules${NODE_PATH:+:$NODE_PATH}"
  fi
fi

rm -rf ./node_modules/.vite-temp ./node_modules/.vite ./test-results ./playwright-report 2>/dev/null || true

exec "$@"
