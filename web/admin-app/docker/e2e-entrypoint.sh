#!/bin/sh
set -e

if [ -d /app/node_modules ] && [ ! -x ./node_modules/.bin/vite ]; then
  mkdir -p ./node_modules
  cp -a /app/node_modules/. ./node_modules/
fi

exec "$@"
