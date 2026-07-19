package webhook

import (
	"context"
	"errors"
	"net/http"
)

var (
	ErrRateLimited  = errors.New("delivery rate limited")
	ErrRejected     = errors.New("delivery rejected")
	ErrUpstream     = errors.New("delivery upstream failed")
	ErrBodyTooLarge = errors.New("delivery response too large")
)

type Command struct {
	Destination string `json:"destination"`
	Message     string `json:"message"`
}
type Result struct {
	DeliveryID string `json:"delivery_id"`
}
type Client struct {
	client  *http.Client
	baseURL string
	maxBody int64
}

func New(client *http.Client, baseURL string, maxBody int64) (*Client, error) {
	return nil, errors.New("TODO: validate the delivery client")
}
func (c *Client) Deliver(ctx context.Context, command Command) (Result, error) {
	return Result{}, errors.New("TODO: deliver the alert")
}
