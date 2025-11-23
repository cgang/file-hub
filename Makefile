BINARY=bin/file-hub
GOFILES=$(shell find . -type f -name "*.go")

$(BINARY): $(GOFILES)
	go build -o $(BINARY) cmd/main.go

build: $(BINARY)

migrate:
	psql -d filehub -f scripts/database_schema.sql

golint:
	golangci-lint run

clean:
	rm -f $(BINARY)

all: build
