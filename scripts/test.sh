#!/bin/sh

set -euo pipefail

echo " * migrating DATABASE_URL"
dockerize \
  -wait tcp://redis:6379 \
  -wait tcp://postgres_test:5432 \
  -timeout 10s \
  -wait-retry-interval 1s \
  migrate -path migrations -database $DATABASE_URL -verbose up

echo " * run tests, watch"
fswatch --config fsw.test.yml
