FROM golang:1.14-alpine

WORKDIR /go/src/app

RUN apk add --no-cache --update \
      openssl\
      git \
      build-base \
      curl \
      postgresql-client

ENV DOCKERIZE_VERSION v0.6.1
RUN wget https://github.com/jwilder/dockerize/releases/download/$DOCKERIZE_VERSION/dockerize-alpine-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
    && tar -C /usr/local/bin -xzvf dockerize-alpine-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
    && rm dockerize-alpine-linux-amd64-$DOCKERIZE_VERSION.tar.gz


ENV MIGRATE_VERSION v4.11.0
RUN go get -u -d github.com/golang-migrate/migrate/cmd/migrate \
      && cd $GOPATH/src/github.com/golang-migrate/migrate/cmd/migrate \
      && git checkout $MIGRATE_VERSION \
      && go build -tags 'postgres' -ldflags="-X main.Version=$(git describe --tags)" -o $GOPATH/bin/migrate $GOPATH/src/github.com/golang-migrate/migrate/cmd/migrate

RUN go get -u golang.org/x/lint/golint
RUN go get github.com/codeskyblue/fswatch
