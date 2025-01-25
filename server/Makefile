.PHONY: build run test clean

# Build the server
build:
	go build -o bin/server cmd/server/main.go

# Run the server
run:
	go run cmd/server/main.go

# Run tests
test:
	go test ./...

# Clean build artifacts
clean:
	rm -rf bin/ 