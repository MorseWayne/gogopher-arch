# syntax=docker/dockerfile:1.7

FROM golang:1.26-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN --mount=type=cache,id=gogopher-go-mod,target=/go/pkg/mod \
    go mod download

COPY cmd/migrate ./cmd/migrate
COPY internal/platform/database ./internal/platform/database
RUN --mount=type=cache,id=gogopher-go-mod,target=/go/pkg/mod \
    --mount=type=cache,id=gogopher-go-build,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -trimpath -o /out/migrate ./cmd/migrate

FROM alpine:3.22

WORKDIR /app
COPY --from=builder /out/migrate /usr/local/bin/migrate
COPY db/migrations ./db/migrations

ENTRYPOINT ["migrate"]
CMD ["up"]
