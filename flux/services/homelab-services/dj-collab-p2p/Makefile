.PHONY: help install dev test build docker-up docker-down clean

help:
	@echo "DJ Collab P2P Game - Makefile"
	@echo ""
	@echo "Available commands:"
	@echo "  make install      - Install all dependencies"
	@echo "  make dev          - Start development environment"
	@echo "  make test         - Run all tests"
	@echo "  make build        - Build all applications"
	@echo "  make docker-up    - Start Docker services"
	@echo "  make docker-down  - Stop Docker services"
	@echo "  make clean        - Clean build artifacts"

install:
	pnpm install
	cd apps/server && go mod download

dev:
	@echo "Starting development environment..."
	@echo "Make sure Docker services are running: make docker-up"
	pnpm dev:server &
	pnpm dev:web &
	@echo "Development servers started"

test:
	pnpm test

build:
	pnpm build
	cd apps/server && go build -o bin/server main.go

docker-up:
	docker-compose up -d
	@echo "Docker services started"

docker-down:
	docker-compose down
	@echo "Docker services stopped"

clean:
	rm -rf node_modules
	rm -rf apps/*/node_modules
	rm -rf packages/*/node_modules
	rm -rf apps/*/dist
	rm -rf packages/*/dist
	rm -rf apps/server/bin
	@echo "Clean complete"
