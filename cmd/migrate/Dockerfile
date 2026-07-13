FROM golang:1.25-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY cmd/migrate ./cmd/migrate
COPY internal/platform/database ./internal/platform/database
RUN CGO_ENABLED=0 go build -trimpath -o /out/migrate ./cmd/migrate

FROM alpine:3.22

WORKDIR /app
COPY --from=builder /out/migrate /usr/local/bin/migrate
COPY db/migrations ./db/migrations

ENTRYPOINT ["migrate"]
CMD ["up"]
