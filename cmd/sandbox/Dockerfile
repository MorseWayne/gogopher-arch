FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /sandbox ./cmd/sandbox

FROM golang:1.25-alpine
RUN addgroup -S sandbox && adduser -S -G sandbox sandbox
COPY --from=builder /sandbox /usr/local/bin/gogopher-sandbox
USER sandbox
ENV CGO_ENABLED=0
EXPOSE 8081
ENTRYPOINT ["/usr/local/bin/gogopher-sandbox"]
