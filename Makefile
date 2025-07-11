SRC_DIR := .
GO := go

# Commands
all: build test

build: fmt
	@echo "Building the binary..."
	$(GO) build -o app $(SRC_DIR)

test:
	@echo "Running tests..."
	$(GO) test ./... -v


fmt:
	@echo "Formatting the code..."
	$(GO) fmt ./...

vet:
	@echo "Vet the code..."
	$(GO) vet ./...

lint:
	@echo "Linting the code..."
	@golint ./...

run:
	@echo "Running the application..."
	go run main.go
exec:
	@podman build --no-cache -f Dockerfile -t api
	@podman run api

docker-build:
	@COMMIT_HASH=$(shell git rev-parse HEAD)
	@podman build --no-cache --build-arg gitsha=$(COMMIT_HASH) -f Dockerfile -t ecr-proxy

push:
	@COMMIT_HASH=$(shell git rev-parse HEAD)
	@podman build --no-cache --build-arg gitsha=$(COMMIT_HASH) -f Dockerfile -t ecr-proxy
	@podman push ecr-proxy $(TARGET)

helm-docs:
	@echo "Generating Helm documentation..."
	@helm-docs
