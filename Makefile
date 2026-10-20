include .env
export

 MIGRATION_DSN = postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DATABASE)?sslmode=disable

build:
	go build -o bin/ cmd/deploy-platform/main.go

run:
	go run cmd/deploy-platform/main.go

lint:
	golangci-lint run

fmt:
	golangci-lint fmt

test:
	go test ./...

migrate-up:
	goose -dir migrations postgres "$(MIGRATION_DSN)" up

migrate-down:
	goose -dir migrations postgres "$(MIGRATION_DSN)" down

migrate-status:
	goose -dir migrations postgres "$(MIGRATION_DSN)" status