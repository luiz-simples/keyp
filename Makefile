.PHONY: build lint format test coverage

build:
	CGO_ENABLED=1 \
		go build \
		-trimpath \
		-ldflags="-s -w" \
		-o ./bin/keyp \
		./cmd/keyp

lint:
	golangci-lint run --fix ./...

test:
	ginkgo -cover ./...

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"
	open coverage.html
