.PHONY: build run test clean docker-build docker-run dev migrate

# Variables
BINARY_NAME=maylng-api
DOCKER_IMAGE=maylng/backend

# Build the application
build:
	go build -o bin/$(BINARY_NAME) ./cmd/api

# Run the application
run:
	go run ./cmd/api

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	go clean
	rm -f bin/$(BINARY_NAME)

# Install dependencies
deps:
	go mod download
	go mod tidy

# Development setup
dev:
	docker-compose up -d postgres redis
	sleep 5
	make migrate-up
	go run ./cmd/api

# Database migrations
migrate-up:
	migrate -path migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path migrations -database "$(DATABASE_URL)" down

migrate-create:
	migrate create -ext sql -dir migrations -seq $(name)

# Docker commands
docker-build:
	docker build -t $(DOCKER_IMAGE) .

docker-run:
	docker-compose up

docker-down:
	docker-compose down

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Install tools
install-tools:
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Security check
security:
	gosec ./...

# Generate API documentation
docs:
	swag init -g cmd/api/main.go
