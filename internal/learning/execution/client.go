package execution

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const defaultMaxSandboxResponseBytes int64 = 4 << 20

var ErrInvalidSandboxResponse = errors.New("invalid Sandbox response")

type SandboxClientOptions struct {
	Endpoint         string
	HTTPClient       *http.Client
	MaxResponseBytes int64
}

type SandboxClient struct {
	endpoint         string
	client           *http.Client
	maxResponseBytes int64
}

func NewSandboxClient(options SandboxClientOptions) (*SandboxClient, error) {
	endpoint, err := url.Parse(options.Endpoint)
	if err != nil || (endpoint.Scheme != "http" && endpoint.Scheme != "https") || endpoint.Host == "" {
		return nil, fmt.Errorf("valid HTTP Sandbox endpoint is required")
	}
	if options.MaxResponseBytes == 0 {
		options.MaxResponseBytes = defaultMaxSandboxResponseBytes
	}
	if options.MaxResponseBytes < 1 {
		return nil, fmt.Errorf("Sandbox response limit must be positive")
	}
	client := options.HTTPClient
	if client == nil {
		client = &http.Client{}
	}
	copy := *client
	copy.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	return &SandboxClient{endpoint: endpoint.String(), client: &copy, maxResponseBytes: options.MaxResponseBytes}, nil
}

func (c *SandboxClient) Execute(ctx context.Context, spec ExecutionSpec) (ExecutionResponse, error) {
	if err := spec.Validate(); err != nil {
		return ExecutionResponse{}, err
	}
	payload, err := json.Marshal(spec)
	if err != nil {
		return ExecutionResponse{}, fmt.Errorf("encode Sandbox request: %w", err)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(payload))
	if err != nil {
		return ExecutionResponse{}, fmt.Errorf("create Sandbox request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := c.client.Do(request)
	if err != nil {
		return ExecutionResponse{}, fmt.Errorf("call Sandbox: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 64<<10))
		return ExecutionResponse{}, fmt.Errorf("Sandbox returned HTTP %d", response.StatusCode)
	}
	responseJSON, err := io.ReadAll(io.LimitReader(response.Body, c.maxResponseBytes+1))
	if err != nil {
		return ExecutionResponse{}, fmt.Errorf("%w: read body: %v", ErrInvalidSandboxResponse, err)
	}
	if int64(len(responseJSON)) > c.maxResponseBytes {
		return ExecutionResponse{}, fmt.Errorf("%w: body exceeds %d bytes", ErrInvalidSandboxResponse, c.maxResponseBytes)
	}
	decoder := json.NewDecoder(bytes.NewReader(responseJSON))
	decoder.DisallowUnknownFields()
	var result ExecutionResponse
	if err := decoder.Decode(&result); err != nil {
		return ExecutionResponse{}, fmt.Errorf("%w: decode JSON: %v", ErrInvalidSandboxResponse, err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return ExecutionResponse{}, fmt.Errorf("%w: trailing data", ErrInvalidSandboxResponse)
	}
	if err := result.Validate(); err != nil {
		return ExecutionResponse{}, fmt.Errorf("%w: %v", ErrInvalidSandboxResponse, err)
	}
	return result, nil
}
