.PHONY: build build-prod clean lint test coverage docker-build

build:
	CGO_ENABLED=1 \
		go build \
		-trimpath \
		-ldflags="-s -w" \
		-o ./bin/keyp \
		./cmd/keyp

docker:
	docker build --platform linux/arm64 -t keyp:latest .

prepare:
	brew tap golangci/tap
	brew install golangci/tap/golangci-lint
	go install go.uber.org/mock/mockgen@latest
	go install github.com/onsi/ginkgo/v2/ginkgo
	go get github.com/onsi/gomega/...
	spctl developer-mode enable-terminal

clean:
	rm -rf ./bin
	rm -f coverage.out coverage.html coverprofile.out service.test

lint:
	golangci-lint run --fix ./...

test:
	ginkgo -cover ./...
# 	ginkgo --randomize-suites --randomize-all -cover ./...

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	open coverage.html

mockgen:
	mockgen -source=${GOPATH}/pkg/mod/github.com/tidwall/redcon@v1.6.2/redcon.go -destination=internal/app/mocks_redcon_test.go -package=app_test
	mockgen -source=internal/domain/types.go -destination=internal/app/mocks_types_test.go -package=app_test
	mockgen -source=internal/domain/types.go -destination=internal/service/mocks_types_test.go -package=service_test

