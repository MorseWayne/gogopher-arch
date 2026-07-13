FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /gateway ./cmd/gateway

FROM alpine:3.22
RUN addgroup -S app && adduser -S -G app app
WORKDIR /app
COPY --from=builder /gateway /app/gateway
COPY --from=builder /app/content/learning /app/content/learning
USER app
EXPOSE 8080
ENTRYPOINT ["/app/gateway"]
