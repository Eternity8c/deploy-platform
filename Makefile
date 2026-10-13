build:
	go build -o bin/ cmd/deploy-platform/main.go

run:
	go run cmd/deploy-platform/main.go

lint:
	golangci-lint run

test:
	go test ./...