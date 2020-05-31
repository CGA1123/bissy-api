#!/bin/sh

echo " * wait for redis and postgres, gwatch"
dockerize \
  -wait tcp://redis:6379 \
  -wait tcp://postgres:5432 \
  -timeout 10s \
  -wait-retry-interval 1s \
  fswatch sh -c "clear; go test ./... -parallel 4"
