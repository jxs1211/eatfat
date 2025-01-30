.PHONY: build run test clean protoc delete-db sqlc-generate

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

delete-db:
	rm -rf db.sqlite

sqlc-generate:
	sqlc generate -f internal/server/db/config/sqlc.yaml