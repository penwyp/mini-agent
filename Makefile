# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOCLEAN=$(GOCMD) clean

# Variables
BINARY_NAME=mini-agent
BINARY_PATH=./bin/$(BINARY_NAME)
MAIN_SRC=./cmd/agent/main.go

# Default target
.DEFAULT_GOAL := all

# Phony targets prevent conflicts with files of the same name
.PHONY: all build run test clean

all: build

# Build the Go application
build:
	@echo "Building the application..."
	@mkdir -p $(dir $(BINARY_PATH))
	@$(GOBUILD) -o $(BINARY_PATH) $(MAIN_SRC)
	@echo "Build complete: $(BINARY_PATH)"

# Run the Go application
# Requires AGENT_API_KEY to be set in the environment.
# Example: export AGENT_API_KEY="your_key" && make run
run: build
	@echo "Running the agent..."
	@AGENT_API_KEY="$(AGENT_API_KEY)" $(BINARY_PATH)

# Run tests
test:
	@echo "Running tests..."
	@$(GOTEST) -v -race ./...

# Clean up build artifacts
clean:
	@echo "Cleaning up..."
	@$(GOCLEAN)
	@rm -f $(BINARY_PATH)
	@if [ -d ./bin ]; then rmdir ./bin 2>/dev/null || true; fi

# Help target
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  all        - Build the application (default)."
	@echo "  build      - Build the application."
	@echo "  run        - Run the built application (requires AGENT_API_KEY)."
	@echo "  test       - Run all tests."
	@echo "  clean      - Clean up build artifacts."
	@echo "  help       - Show this help message." 