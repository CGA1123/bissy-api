#!/bin/sh

set -euo pipefail

echo " * migrating DATABASE_URL=$DATABASE_URL"
migrate -path migrations -database $DATABASE_URL -verbose up

echo " * wait for redis and postgres, gwatch"
dockerize \
  -wait tcp://redis:6379 \
  -wait tcp://postgres_test:5432 \
  -timeout 10s \
  -wait-retry-interval 1s \
  fswatch --config fsw.test.yml
