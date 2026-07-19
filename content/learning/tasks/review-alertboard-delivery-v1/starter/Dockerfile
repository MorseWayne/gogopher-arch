FROM golang:1.26-alpine AS builder
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go test ./... && CGO_ENABLED=0 go build -trimpath -o /out/alertboard ./cmd/alertboard

FROM alpine:3.23
RUN addgroup -S app && adduser -S -G app app
COPY --from=builder /out/alertboard /usr/local/bin/alertboard
USER app
ENTRYPOINT ["/usr/local/bin/alertboard"]
