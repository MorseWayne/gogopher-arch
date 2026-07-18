# syntax=docker/dockerfile:1.7

FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN --mount=type=cache,id=gogopher-go-mod,target=/go/pkg/mod \
    go mod download
COPY . .
RUN --mount=type=cache,id=gogopher-go-mod,target=/go/pkg/mod \
    --mount=type=cache,id=gogopher-go-build,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -o /sandbox ./cmd/sandbox

FROM golang:1.25-alpine
RUN apk add --no-cache gcc musl-dev
RUN addgroup -S sandbox && adduser -S -G sandbox sandbox
COPY --from=builder /sandbox /usr/local/bin/gogopher-sandbox
USER sandbox
ENV CGO_ENABLED=0
EXPOSE 8081
ENTRYPOINT ["/usr/local/bin/gogopher-sandbox"]
