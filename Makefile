.PHONY: build run test clean migrate-up migrate-down docker-up docker-down

# Build the application
build:
	@echo "Building application..."
	@go build -o bin/server cmd/server/main.go
	@echo "Build complete: bin/server"

# Run the application
run:
	@echo "Running application..."
	@go run cmd/server/main.go

# Run tests
test:
	@echo "Running tests..."
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy
	@echo "Dependencies installed"

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@echo "Format complete"

# Lint code
lint:
	@echo "Linting code..."
	@golangci-lint run ./...

# Create database
db-create:
	@echo "Creating database..."
	@mysql -u root -p -e "CREATE DATABASE IF NOT EXISTS go_backend CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
	@echo "Database created"

# Drop database
db-drop:
	@echo "Dropping database..."
	@mysql -u root -p -e "DROP DATABASE IF EXISTS go_backend;"
	@echo "Database dropped"

# Start docker services
docker-up:
	@echo "Starting Docker services..."
	@docker-compose up -d
	@echo "Docker services started"

# Stop docker services
docker-down:
	@echo "Stopping Docker services..."
	@docker-compose down
	@echo "Docker services stopped"

# Show help
help:
	@echo "Available commands:"
	@echo "  make build       - Build the application"
	@echo "  make run         - Run the application"
	@echo "  make test        - Run tests with coverage"
	@echo "  make clean       - Clean build artifacts"
	@echo "  make deps        - Install dependencies"
	@echo "  make fmt         - Format code"
	@echo "  make lint        - Lint code"
	@echo "  make db-create   - Create database"
	@echo "  make db-drop     - Drop database"
	@echo "  make docker-up   - Start Docker services"
	@echo "  make docker-down - Stop Docker services"
