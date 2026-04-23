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
    util-linux-misc \
    net-tools

RUN addgroup -S netdash && adduser -S -u 10001 -G netdash netdash

COPY --from=builder /out/netdash /app/netdash
COPY --from=builder /src/static /app/static

RUN chown -R netdash:netdash /app

EXPOSE 8080
# Container starts as root so setpriv can promote net_raw into the ambient
# set before dropping to uid/gid 10001. This is more reliable than setcap
# because it does not depend on xattr support in the overlay filesystem.
ENTRYPOINT ["setpriv", "--inh-caps=+net_raw", "--ambient-caps=+net_raw", "--reuid=10001", "--regid=10001", "--init-groups", "/app/netdash"]
