BINARY=bin/file-hub
GOFILES=$(shell find . -type f -name "*.go")
JSFILES=$(shell find web/src -type f -name "*.js" -o -name "*.svelte" -o -name "*.css")
PROTOFILES=$(shell find . -type f -name "*.proto")

# Generate protobuf Go files before building
$(BINARY): $(GOFILES) $(JSFILES) pkg/sync/sync.pb.go
	# Install frontend dependencies and build assets
	cd web && npm install && npm run build
	# Then build the Go binary with embedded assets
	go build -o $(BINARY) cmd/main.go

build: $(BINARY)

# Rule to generate protobuf Go files
pkg/sync/sync.pb.go: pkg/sync/sync.proto
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		pkg/sync/sync.proto


golint:
	golangci-lint run

clean:
	rm -f $(BINARY)
	rm -rf dist
	rm -rf web/dist
	rm -rf node_modules
	rm -rf web/node_modules
	find . -name "*.pb.go" -delete

all: build

.PHONY: clean all
