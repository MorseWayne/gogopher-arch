FROM golang:1.25-alpine
COPY . /src
RUN cd /src && go build -o /alertworker ./cmd/alertworker
ENTRYPOINT ["/alertworker"]
