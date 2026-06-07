#!/bin/sh
set -e

if [ -d /app/node_modules ] && [ ! -x ./node_modules/.bin/vite ]; then
  if ! mkdir -p ./node_modules 2>/dev/null || ! cp -R /app/node_modules/. ./node_modules/ 2>/dev/null; then
    echo "workspace node_modules is not writable; using image layer dependencies"
    export PATH="/app/node_modules/.bin:$PATH"
    export NODE_PATH="/app/node_modules${NODE_PATH:+:$NODE_PATH}"
  fi
fi

rm -rf ./node_modules/.vite-temp ./test-results ./playwright-report 2>/dev/null || true

exec "$@"
