FROM golang:1.22-alpine as builder

WORKDIR /app
COPY go.mod ./

COPY . .
RUN go build -o main ./src/services/sandbox-engine/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/main .
CMD ["./main"]
