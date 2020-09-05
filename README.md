# :robot: bissy-api

![CI](https://github.com/CGA1123/bissy-api/workflows/CI/badge.svg)
[![Maintainability](https://api.codeclimate.com/v1/badges/d42cd3823b699de16259/maintainability)](https://codeclimate.com/github/CGA1123/bissy-api/maintainability)
[![Test Coverage](https://api.codeclimate.com/v1/badges/d42cd3823b699de16259/test_coverage)](https://codeclimate.com/github/CGA1123/bissy-api/test_coverage)
[![Go Report Card](https://goreportcard.com/badge/github.com/CGA1123/bissy-api)](https://goreportcard.com/report/github.com/CGA1123/bissy-api)

Some toy APIs to learn go!

## Getting Started

This project comes with a `docker-compose` setup! If you want to run the webserver use `docker-compose up dev`. The webserver will watch `*.go` files and recompile and restart whenever you make a change.

For testing, run `docker-compose up test`, again this process will watch for changes to `*.go` files and recompile and re-run the test suite on every change.

## Auth

Authentication is done via JWT Token or API Token.
You can get a JWT token via Github OAuth, via this url:
```
https://api.bissy.io/auth/github/signin?redirect_uri=https://api.bissy.io/auth/github/token
```

Subsequent request to the API will need to set the `Authorization` header with the `Bearer` token equal to the token returned from the above request. A potential exchange might look like this (on macOS):

```
open "https://api.bissy.io/auth/github/signin?redirect_uri=https://api.bissy.io/auth/github/token"
# Returns a JSON payload of `{ "token": "a-jwt-token" }

curl -i -H "Authorization: Bearer a-jwt-token" "https://api.bissy.io/authping
```

To create an apikey:

```
curl -i -H "Authorization: Bearer a-jwt-token" \
        -H "Content-Type: application/json" \
        -d '{ "name": "Personal API Key" }' \
        "https://api.bissy.io/auth/apikeys"

# Returns { id: "an-id", "name": "Personal API Key", "key": "the-api-key"}
# The "key" value will no longer be exposed after this call, make sure you keep it safe!

curl -i -H "X-Bissy-Apikey: the-api-key" "https://api.bissy.io/authping"
```
