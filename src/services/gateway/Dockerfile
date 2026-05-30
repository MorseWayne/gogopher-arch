FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod ./
# Skip go sum for now as we don't have many dependencies yet
# RUN go mod download

COPY . .
RUN go build -o main ./src/services/gateway/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/main .
EXPOSE 8080
CMD ["./main"]
