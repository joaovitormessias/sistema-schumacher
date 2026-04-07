#!/bin/sh
set -eu

if [ -n "${ENV_FILE:-}" ] && [ -f "${ENV_FILE}" ]; then
  set -a
  . "${ENV_FILE}"
  set +a
elif [ -f "/app/.env" ]; then
  set -a
  . /app/.env
  set +a
fi

exec /app/schumacher-api
