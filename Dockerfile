FROM golang:1.25-alpine AS builder

RUN apk add --no-cache \
    gcc \
    musl-dev \
    lmdb-dev \
    git \
    ca-certificates

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

RUN CGO_ENABLED=1 \
    GOARCH=arm64 \
    go build \
        -a \
        -installsuffix cgo \
        -trimpath \
        -ldflags="-s -w -extldflags '-static'" \
        -tags netgo \
        -o keyp \
        ./cmd/keyp

FROM alpine:latest

RUN apk add --no-cache \
    lmdb \
    ca-certificates \
    tzdata && \
    addgroup -g 1001 -S keyp && \
    adduser -u 1001 -S keyp -G keyp

RUN mkdir -p /data && \
    chown -R keyp:keyp /data

COPY --from=builder /build/keyp /usr/local/bin/keyp

USER keyp
EXPOSE 6379
VOLUME ["/data"]

ENV KEYP_ADDRESS=0.0.0.0:6379
ENV KEYP_DATA_DIR=/data

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD nc -z localhost 6379 || exit 1

ENTRYPOINT ["/usr/local/bin/keyp"]
