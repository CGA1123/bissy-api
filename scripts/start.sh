#!/bin/sh

set -euo pipefail

echo " * migrating DATABASE_URL"
dockerize \
  -wait tcp://redis:6379 \
  -wait tcp://postgres:5432 \
  -timeout 10s \
  -wait-retry-interval 1s \
  migrate -path migrations -database $DATABASE_URL -verbose up

echo " * start server, and watch..."
fswatch --config fsw.app.yml
