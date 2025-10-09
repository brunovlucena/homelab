.PHONY: help build run test clean docker-build docker-run dev tidy lint

# 🎯 Default target
help:
	@echo "🌐 Bruno Site API - Available commands:"
	@echo ""
	@echo "  make build        - Build the API binary"
	@echo "  make run          - Run the API server"
	@echo "  make dev          - Run with hot reload (requires air)"
	@echo "  make test         - Run tests"
	@echo "  make tidy         - Tidy and download dependencies"
	@echo "  make lint         - Run linter"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make docker-build - Build Docker image"
	@echo "  make docker-run   - Run Docker container"
	@echo ""

# 🔨 Build the binary
build:
	@echo "🔨 Building API..."
	go build -o bin/bruno-site-api .
	@echo "✅ Build complete: bin/bruno-site-api"

# 🚀 Run the API
run:
	@echo "🚀 Starting API..."
	go run main.go

# 🔥 Run with hot reload
dev:
	@echo "🔥 Starting API with hot reload..."
	@which air > /dev/null || (echo "Installing air..." && go install github.com/cosmtrek/air@latest)
	air

# 🧪 Run tests
test:
	@echo "🧪 Running tests..."
	go test -v ./...

# 📦 Tidy dependencies
tidy:
	@echo "📦 Tidying dependencies..."
	go mod tidy
	@echo "✅ Dependencies updated"

# 🔍 Run linter
lint:
	@echo "🔍 Running linter..."
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run

# 🧹 Clean build artifacts
clean:
	@echo "🧹 Cleaning..."
	rm -rf bin/
	rm -rf tmp/
	rm -f build-errors.log
	@echo "✅ Clean complete"

# 🐳 Build Docker image
docker-build:
	@echo "🐳 Building Docker image..."
	docker build -t bruno-site-api:latest .
	@echo "✅ Docker image built: bruno-site-api:latest"

# 🐳 Run Docker container
docker-run:
	@echo "🐳 Running Docker container..."
	docker run -p 8080:8080 \
		-e DATABASE_URL=${DATABASE_URL} \
		-e REDIS_URL=${REDIS_URL} \
		-e MINIO_ENDPOINT=${MINIO_ENDPOINT} \
		bruno-site-api:latest

