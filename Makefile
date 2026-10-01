APP_NAME=real-estate-backend

.PHONY: run build test test-race lint fmt vet tidy swag docker-up docker-down

run:
	go run ./cmd/...

build:
	go build -o bin/$(APP_NAME) ./cmd/...

test:
	go test ./... -count=1

test-race:
	go test ./... -race -count=1

lint:
	golangci-lint run ./...

fmt:
	gofmt -w .

vet:
	go vet ./...

tidy:
	go mod tidy

swag:
	swag init -g cmd/main.go -o docs

docker-up:
	docker compose up -d

docker-down:
	docker compose down
