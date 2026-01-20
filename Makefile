.PHONY: build test clean install lint

# Build the bearing binary
build:
	go build -o bearing ./cmd/bearing

# Install to GOPATH/bin
install:
	go install ./cmd/bearing

# Run unit tests
test:
	BEARING_AI_ENABLED=0 go test ./...

# Run integration tests (requires built binary)
integration: build
	chmod +x test/run_tests.sh
	PATH="$(PWD):$(PATH)" BEARING_AI_ENABLED=0 ./test/run_tests.sh

# Run all tests
test-all: test integration

# Clean build artifacts
clean:
	rm -f bearing

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	go vet ./...

# Tidy dependencies
tidy:
	go mod tidy
