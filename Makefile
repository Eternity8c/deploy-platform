include .env
export

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