package httpclient

import (
	"context"
	"errors"
	"net/http"
)

var (
	ErrRateLimited  = errors.New("probe rate limited")
	ErrRejected     = errors.New("probe request rejected")
	ErrUpstream     = errors.New("probe upstream failed")
	ErrBodyTooLarge = errors.New("probe response too large")
)

type Result struct {
	Status string `json:"status"`
}
type Client struct {
	client  *http.Client
	baseURL string
	maxBody int64
}

func New(client *http.Client, baseURL string, maxBody int64) (*Client, error) {
	return nil, errors.New("TODO: validate the probe client")
}
func (c *Client) Probe(ctx context.Context, target string) (Result, error) {
	return Result{}, errors.New("TODO: call the external probe")
}
