.PHONY: start build test lint docker-up docker-down

BIN := bin/api

start:
	go run ./cmd/api

build:
	go build -o $(BIN) ./cmd/api

test:
	go test ./... -v -race -count=1

lint:
	golangci-lint run ./...

docker-up:
	docker compose up -d

docker-down:
	docker compose down
