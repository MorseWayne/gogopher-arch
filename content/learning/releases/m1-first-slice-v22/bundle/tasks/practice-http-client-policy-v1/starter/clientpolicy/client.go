package clientpolicy

import (
	"errors"
	"net/http"
	"time"
)

type Config struct {
	Timeout               time.Duration
	MaxIdleConns          int
	MaxIdleConnsPerHost   int
	IdleConnTimeout       time.Duration
	TLSHandshakeTimeout   time.Duration
	ResponseHeaderTimeout time.Duration
}

func NewClient(config Config) (*http.Client, error) {
	return nil, errors.New("TODO: build an isolated HTTP client")
}
