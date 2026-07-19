FROM golang:1.25-alpine
WORKDIR /src
COPY . .
RUN go build -o /gocheckhub ./cmd/gocheckhub
ENTRYPOINT ["/gocheckhub"]
