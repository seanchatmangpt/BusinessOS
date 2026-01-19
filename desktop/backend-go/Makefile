.PHONY: help build test voice-test voice-test-quick voice-test-audio clean

# Default target
help:
	@echo "BusinessOS Backend - Available Commands"
	@echo ""
	@echo "Build & Run:"
	@echo "  make build              Build the server binary"
	@echo "  make run                Run the server"
	@echo "  make test               Run unit tests"
	@echo ""
	@echo "Voice Testing:"
	@echo "  make voice-test         Run full voice E2E test suite"
	@echo "  make voice-test-quick   Run quick voice tests (basic cases)"
	@echo "  make voice-test-audio   Generate test audio files"
	@echo "  make voice-test-clean   Clean voice test results"
	@echo ""
	@echo "Database:"
	@echo "  make migrate            Run database migrations"
	@echo "  make sqlc               Generate SQLC code"
	@echo ""
	@echo "Utilities:"
	@echo "  make clean              Clean build artifacts"
	@echo "  make lint               Run linters"
	@echo ""

# Build
build:
	@echo "Building server..."
	go build -o bin/server ./cmd/server

# Run
run:
	@echo "Starting server..."
	go run ./cmd/server

# Test
test:
	@echo "Running tests..."
	go test ./...

# Voice Testing
voice-test:
	@echo "Running voice E2E tests..."
	./scripts/test_voice_e2e.sh

voice-test-quick:
	@echo "Running quick voice tests..."
	./scripts/test_voice_e2e.sh --quick

voice-test-audio:
	@echo "Generating test audio files..."
	./scripts/generate_test_audio.sh

voice-test-verbose:
	@echo "Running voice tests (verbose)..."
	./scripts/test_voice_e2e.sh --verbose

voice-test-clean:
	@echo "Cleaning voice test results..."
	rm -rf test-results/voice/*
	@echo "Voice test results cleaned"

# Database
migrate:
	@echo "Running migrations..."
	go run ./cmd/migrate

sqlc:
	@echo "Generating SQLC code..."
	sqlc generate

# Utilities
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -rf test-results/
	@echo "Clean complete"

lint:
	@echo "Running linters..."
	golangci-lint run ./...
