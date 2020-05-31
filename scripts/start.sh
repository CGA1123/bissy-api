#!/bin/sh

echo " * go get"
go get -d -v ./...

echo " * go install"
go install -v ./...

echo " * wait for redis and postgres, start server"
dockerize \
  -wait tcp://redis:6379 \
  -wait tcp://postgres:5432 \
  -timeout 10s \
  -wait-retry-interval 1s \
  bissy-api
