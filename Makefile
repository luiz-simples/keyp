.PHONY: build build-prod clean lint test coverage docker-build

build:
	CGO_ENABLED=1 \
		go build \
		-trimpath \
		-ldflags="-s -w" \
		-o ./bin/keyp \
		./cmd/keyp

build-prod:
	CGO_ENABLED=1 \
		go build \
		-a \
		-installsuffix cgo \
		-trimpath \
		-ldflags="-s -w" \
		-tags netgo \
		-o ./bin/keyp \
		./cmd/keyp

docker-build:
	docker build --platform linux/arm64 -t keyp:latest .

clean:
	rm -rf ./bin
	rm -f coverage.out coverage.html coverprofile.out service.test

lint:
	golangci-lint run --fix ./...

test:
	ginkgo -cover ./...

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	open coverage.html
