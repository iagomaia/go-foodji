.PHONY: start build test lint docker-up docker-down swag

BIN := bin/api
SWAG := $(shell go env GOPATH)/bin/swag

start:
	go run ./cmd/api

build:
	go build -o $(BIN) ./cmd/api

test:
	go test ./... -v -race -count=1

swag:
	$(SWAG) init -g cmd/api/main.go --output docs --outputTypes go,yaml

docker-up:
	docker compose up -d

docker-down:
	docker compose down
