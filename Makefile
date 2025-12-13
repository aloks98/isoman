.PHONY: help build run stop clean test dev-backend dev-frontend docker-build docker-run docker-stop docker-clean

# Default target
help:
	@echo "ISOMan - Linux ISO Download Manager"
	@echo ""
	@echo "Available targets:"
	@echo "  make dev-backend       - Run backend in development mode"
	@echo "  make dev-frontend      - Run frontend in development mode"
	@echo "  make build            - Build frontend and backend"
	@echo "  make docker-build     - Build Docker image"
	@echo "  make docker-run       - Run with docker-compose"
	@echo "  make docker-stop      - Stop docker-compose"
	@echo "  make docker-clean     - Remove Docker containers and volumes"
	@echo "  make clean            - Clean build artifacts"
	@echo "  make test             - Run tests"

# Development
dev-backend:
	@echo "Starting backend server..."
	cd backend && go run main.go

dev-frontend:
	@echo "Starting frontend dev server..."
	cd ui && bun run dev

# Build
build: build-frontend build-backend

build-frontend:
	@echo "Building frontend..."
	cd ui && bun install && bun run build

build-backend:
	@echo "Building backend..."
	cd backend && go build -o server .

# Docker
docker-build:
	@echo "Building Docker image..."
	docker build -t isoman:latest .

docker-run:
	@echo "Starting ISOMan with docker-compose..."
	docker-compose up -d

docker-stop:
	@echo "Stopping ISOMan..."
	docker-compose down

docker-clean:
	@echo "Cleaning up Docker resources..."
	docker-compose down -v
	docker rmi isoman:latest 2>/dev/null || true

# Testing
test:
	@echo "Running backend tests..."
	cd backend && go test ./...

# Clean
clean:
	@echo "Cleaning build artifacts..."
	rm -rf ui/dist
	rm -f backend/server
	rm -rf backend/data
