# syntax=docker/dockerfile:1

FROM golang:1.26-alpine AS builder
WORKDIR /log_collect

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /app ./cmd/api/main.go


FROM alpine:3.24
RUN apk add --no-cache ca-certificates tzdata \
 && adduser -D -u 10001 app

ENV APP_ENV=production \
    GIN_MODE=release

WORKDIR /app
COPY --from=builder /app ./app

USER app
CMD ["./app"]