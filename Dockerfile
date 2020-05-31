FROM golang:1.14-alpine

WORKDIR /go/src/app
COPY . .

RUN apk add --no-cache openssl

ENV DOCKERIZE_VERSION v0.6.1
RUN wget https://github.com/jwilder/dockerize/releases/download/$DOCKERIZE_VERSION/dockerize-alpine-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
    && tar -C /usr/local/bin -xzvf dockerize-alpine-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
    && rm dockerize-alpine-linux-amd64-$DOCKERIZE_VERSION.tar.gz

RUN go get -d -v ./...
RUN go install -v ./...

HEALTHCHECK --interval=5m --timeout=3s \
  CMD curl -f http://localhost:8080/ping || exit 1
