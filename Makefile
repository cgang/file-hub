BINARY=bin/file-hub
GOFILES=$(shell find . -type f -name "*.go")

$(BINARY): $(GOFILES)
	go build -o $(BINARY) cmd/main.go

build: $(BINARY)

golint:
	golangci-lint run

clean:
	rm -f $(BINARY)

all: build
