# Makefile

# Variables
BINARY_NAME=server
GO_FILES=main.go
BUILD_DIR=bin
GOFMT_FILES=$(shell find . -name '*.go' -not -path './vendor/*')

# Targets

# Default target
all: build

# Build the binary
build:
	@echo "Building the Go server..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(GO_FILES)
	@echo "Build complete. Binary is located in $(BUILD_DIR)/$(BINARY_NAME)."

# Run the server
run: build
	@echo "Running the Go server..."
	@$(BUILD_DIR)/$(BINARY_NAME)

# Clean build artifacts
clean:
	@echo "Cleaning up..."
	rm -rf $(BUILD_DIR)
	@echo "Cleanup complete."

# Format Go files
fmt:
	@echo "Formatting Go files..."
	go fmt $(GOFMT_FILES)
	@echo "Formatting complete."

# Lint Go files (requires golint to be installed)
lint:
	@echo "Linting Go files..."
	golint $(GO_FILES)
	@echo "Linting complete."

# Help
help:
	@echo "Makefile commands:"
	@echo "  make build   - Build the Go server binary."
	@echo "  make run     - Build and run the Go server."
	@echo "  make clean   - Remove build artifacts."
	@echo "  make fmt     - Format Go source files."
	@echo "  make lint    - Lint Go source files."
	@echo "  make help    - Display this help message."

.PHONY: all build run clean fmt lint help