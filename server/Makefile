.PHONY: build run test clean protoc

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

# Run the protoc to generate the go code from the proto file
protoc:
	protoc -I="shared" --go_out="server" "shared/packets.proto"