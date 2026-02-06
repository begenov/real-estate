APP_NAME=real-estate-backend

.PHONY: run build test lint docker-up docker-down

run:
	go run ./cmd

build:
	go build -o bin/$(APP_NAME) ./cmd

test:
	go test ./...

lint:
	golangci-lint run

docker-up:
	docker compose up -d

docker-down:
	docker compose down
