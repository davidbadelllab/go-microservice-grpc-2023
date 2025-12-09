.PHONY: proto build run test clean docker

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=server
CLIENT_NAME=client

# Proto parameters
PROTO_DIR=api/proto
PROTO_OUT=proto

# Build the project
build:
	$(GOBUILD) -o bin/$(BINARY_NAME) ./cmd/server
	$(GOBUILD) -o bin/$(CLIENT_NAME) ./cmd/client

# Run the server
run:
	$(GOCMD) run ./cmd/server/main.go

# Run the client
run-client:
	$(GOCMD) run ./cmd/client/main.go

# Run tests
test:
	$(GOTEST) -v ./...

# Run tests with coverage
test-coverage:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Run integration tests
test-integration:
	$(GOTEST) -v -tags=integration ./...

# Generate proto files
proto:
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		$(PROTO_DIR)/user.proto

# Install proto tools
proto-tools:
	$(GOGET) google.golang.org/protobuf/cmd/protoc-gen-go@latest
	$(GOGET) google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Download dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

# Docker commands
docker-build:
	docker build -t grpc-microservice .

docker-run:
	docker run -p 50051:50051 -p 9090:9090 grpc-microservice

docker-compose-up:
	docker-compose up --build

docker-compose-down:
	docker-compose down -v

# Linting
lint:
	golangci-lint run ./...

# Format code
fmt:
	$(GOCMD) fmt ./...

# Load test with ghz
load-test:
	ghz --insecure --proto $(PROTO_DIR)/user.proto \
		--call user.UserService/CreateUser \
		-d '{"email":"test@example.com","name":"Test User"}' \
		-n 1000 -c 10 \
		localhost:50051

# Database migrations
migrate-up:
	migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/users?sslmode=disable" up

migrate-down:
	migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/users?sslmode=disable" down

# Help
help:
	@echo "Available commands:"
	@echo "  make build          - Build server and client binaries"
	@echo "  make run            - Run the gRPC server"
	@echo "  make run-client     - Run the gRPC client"
	@echo "  make test           - Run unit tests"
	@echo "  make test-coverage  - Run tests with coverage report"
	@echo "  make proto          - Generate protobuf files"
	@echo "  make proto-tools    - Install protobuf tools"
	@echo "  make deps           - Download dependencies"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make docker-build   - Build Docker image"
	@echo "  make docker-run     - Run Docker container"
	@echo "  make docker-compose-up   - Start all services with docker-compose"
	@echo "  make docker-compose-down - Stop all services"
	@echo "  make lint           - Run linter"
	@echo "  make fmt            - Format code"
	@echo "  make load-test      - Run load test with ghz"
