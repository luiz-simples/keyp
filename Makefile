.PHONY: build test test-unit test-integration test-property test-coverage benchmark clean run run-dev deps lint format all docker-build docker-run

build:
	CGO_ENABLED=1 \
		go build \
		-trimpath \
		-ldflags="-s -w" \
		-o ./bin/keyp \
		./cmd/keyp

test: test-unit test-integration test-property

test-unit:
	go test ./internal/storage -v -run "TestStorage"

test-integration:
	go test ./internal/server -v

test-property:
	go test ./internal/storage -v -run "Property"

test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

benchmark:
	go test ./internal/storage -bench=. -benchmem

clean:
	rm -rf ./bin
	rm -rf $(DATA_DIR)
	rm -rf /tmp/keyp-*
	rm -rf /tmp/lmdb-*
	rm -f coverage.out coverage.html

run: build
	././bin/keyp

run-dev:
	go run ./cmd/keyp

deps:
	go mod tidy
	go mod download

lint:
	golangci-lint run

format:
	go fmt ./...

all: clean deps format build test

docker-build:
	docker build -t keyp .

docker-run:
	docker run -p 6380:6380 keyp
