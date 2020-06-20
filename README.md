# :robot: bissy-api

![CI](https://github.com/CGA1123/bissy-api/workflows/CI/badge.svg)
[![Maintainability](https://api.codeclimate.com/v1/badges/d42cd3823b699de16259/maintainability)](https://codeclimate.com/github/CGA1123/bissy-api/maintainability)
[![Test Coverage](https://api.codeclimate.com/v1/badges/d42cd3823b699de16259/test_coverage)](https://codeclimate.com/github/CGA1123/bissy-api/test_coverage)
[![Go Report Card](https://goreportcard.com/badge/github.com/CGA1123/bissy-api)](https://goreportcard.com/report/github.com/CGA1123/bissy-api)

Some toy APIs to learn go!

## Getting Started

This project comes with a `docker-compose` setup! If you want to run the webserver use `docker-compose up dev`. The webserver will watch `*.go` files and recompile and restart whenever you make a change.

For testing, run `docker-compose up test`, again this process will watch for changes to `*.go` files and recompile and re-run the test suite on every change.
