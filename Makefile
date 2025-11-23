BINARY=bin/file-hub
GOFILES=$(shell find . -type f -name "*.go")
JSFILES=$(shell find web/src -type f -name "*.js" -o -name "*.svelte" -o -name "*.css")

$(BINARY): $(GOFILES) $(JSFILES)
	# Install frontend dependencies and build assets first
	cd web && npm install && npm run build
	# Then build the Go binary with embedded assets
	go build -o $(BINARY) cmd/main.go

build: $(BINARY)

run: build
	CONFIG_PATH=example ./$(BINARY)

migrate:
	psql -d filehub -f scripts/database_schema.sql

# Web UI targets
web-dev:
	cd web && npm run dev

web-build:
	cd web && npm run build

web-install:
	cd web && npm install

web-serve:
	cd web && npm run preview

golint:
	golangci-lint run

clean:
	rm -f $(BINARY)
	rm -rf dist
	rm -rf web/dist
	rm -rf node_modules
	rm -rf web/node_modules

all: build

.PHONY: web-dev web-build web-install web-serve clean all run
