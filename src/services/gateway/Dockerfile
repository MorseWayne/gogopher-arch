FROM golang:1.22-alpine as builder

WORKDIR /app
COPY go.mod ./
# Skip go sum for now as we don't have many dependencies yet
# RUN go mod download

COPY . .
RUN go build -o main ./src/services/gateway/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/main .
CMD ["./main"]
