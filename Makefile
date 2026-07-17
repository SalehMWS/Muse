.PHONY: run build test lint fmt vet tidy sqlc-generate migrate-up migrate-down migrate-status docker-up docker-down docker-build

APP_BIN := bin/api

run:
	go run ./cmd/api

build:
	CGO_ENABLED=0 go build -o $(APP_BIN) ./cmd/api

test:
	go test -race -count=1 ./...

lint:
	golangci-lint run ./...

fmt:
	gofmt -l -w .
	goimports -l -w .

vet:
	go vet ./...

tidy:
	go mod tidy

sqlc-generate:
	sqlc generate

migrate-up:
	go run ./cmd/migrate up

migrate-down:
	go run ./cmd/migrate down

migrate-status:
	go run ./cmd/migrate status

docker-up:
	docker compose -f deployments/docker-compose.yml up -d

docker-down:
	docker compose -f deployments/docker-compose.yml down

docker-build:
	docker compose -f deployments/docker-compose.yml build
