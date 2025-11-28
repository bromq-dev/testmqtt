.PHONY: all build clean test conformance-v3 conformance-v5 broker-up broker-down help

# Variables
BINARY_NAME=testmqtt
BIN_DIR=bin
MAIN_PATH=.
BROKER_URL=tcp://localhost:1883

.DEFAULT_GOAL := help

all: build ## Build the binary (default target)

build: ## Build the binary
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete: $(BIN_DIR)/$(BINARY_NAME)"

clean: ## Remove build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf $(BIN_DIR)
	@echo "Clean complete"

test: ## Run Go unit tests
	@echo "Running tests..."
	go test ./...

conformance-v3: build ## Run MQTT v3.1.1 conformance tests
	@echo "Running MQTT v3.1.1 conformance tests..."
	./$(BIN_DIR)/$(BINARY_NAME) conformance --version 3 --broker $(BROKER_URL)

conformance-v3-verbose: build ## Run MQTT v3.1.1 conformance tests (verbose)
	@echo "Running MQTT v3.1.1 conformance tests (verbose)..."
	./$(BIN_DIR)/$(BINARY_NAME) conformance --version 3 --broker $(BROKER_URL) --verbose

conformance-v5: build ## Run MQTT v5.0 conformance tests
	@echo "Running MQTT v5.0 conformance tests..."
	./$(BIN_DIR)/$(BINARY_NAME) conformance --version 5 --broker $(BROKER_URL)

conformance-v5-verbose: build ## Run MQTT v5.0 conformance tests (verbose)
	@echo "Running MQTT v5.0 conformance tests (verbose)..."
	./$(BIN_DIR)/$(BINARY_NAME) conformance --version 5 --broker $(BROKER_URL) --verbose

conformance: conformance-v3 conformance-v5 ## Run all conformance tests

broker-up: ## Start local Mosquitto broker
	@echo "Starting Mosquitto broker..."
	docker compose up -d
	@echo "Broker started on $(BROKER_URL)"

broker-down: ## Stop local Mosquitto broker
	@echo "Stopping Mosquitto broker..."
	docker compose down
	@echo "Broker stopped"

broker-logs: ## Show broker logs
	docker compose logs -f

deps: ## Install dependencies
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

fmt: ## Format code
	@echo "Formatting code..."
	go fmt ./...

lint: ## Run linter
	@echo "Running linter..."
	golangci-lint run ./...

help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
