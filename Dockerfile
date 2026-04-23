# syntax=docker/dockerfile:1.7

FROM golang:1.26-alpine AS builder
WORKDIR /src

# SQLite3 requires CGO toolchain during build.
RUN apk add --no-cache \
    build-base \
    pkgconf \
    sqlite-dev

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal
COPY static ./static

RUN CGO_ENABLED=1 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/netdash ./cmd/server

FROM alpine:3.22 AS runtime
WORKDIR /app

RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    sqlite-libs \
    net-tools

COPY --from=builder /out/netdash /app/netdash
COPY --from=builder /src/static /app/static

EXPOSE 8080
ENTRYPOINT ["/app/netdash"]
