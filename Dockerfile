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
    libcap-utils \
    net-tools

RUN addgroup -S netdash && adduser -S -u 10001 -G netdash netdash

COPY --from=builder /out/netdash /app/netdash
COPY --from=builder /src/static /app/static

RUN chown -R netdash:netdash /app

EXPOSE 8080
# capsh (from libcap-utils) raises net_raw into the inheritable+ambient sets
# before dropping to uid 10001 via --user. It internally uses PR_SET_KEEPCAPS
# so the capability survives the uid transition. More reliable than setcap
# because it does not depend on xattr preservation in the overlay filesystem.
ENTRYPOINT ["capsh", "--caps=cap_net_raw+eip", "--user=netdash", "--addamb=cap_net_raw", "--exec=/app/netdash"]
