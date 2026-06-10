BINARY_NAME := cortex-git-server
BIN_DIR     := bin
MODULE      := github.com/LucasSymons/Cortex/mcp/git-server

.PHONY: fmt test validate build build-all clean

fmt:
	gofmt -w .

test:
	go test ./...

validate:
	go vet ./...
	go build ./...
	go test ./...

build:
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(BINARY_NAME) ./cmd/server

build-all:
	mkdir -p $(BIN_DIR)
	GOOS=darwin  GOARCH=amd64  go build -o $(BIN_DIR)/$(BINARY_NAME)-darwin-amd64  ./cmd/server
	GOOS=darwin  GOARCH=arm64  go build -o $(BIN_DIR)/$(BINARY_NAME)-darwin-arm64  ./cmd/server
	GOOS=linux   GOARCH=amd64  go build -o $(BIN_DIR)/$(BINARY_NAME)-linux-amd64   ./cmd/server
	GOOS=windows GOARCH=amd64  go build -o $(BIN_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/server

clean:
	rm -rf $(BIN_DIR)
