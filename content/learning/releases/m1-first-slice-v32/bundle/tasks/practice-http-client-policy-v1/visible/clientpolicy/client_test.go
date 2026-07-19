package clientpolicy

import (
	"net/http"
	"testing"
	"time"
)

func TestNewClientPolicy(t *testing.T) {
	config := Config{Timeout: 3 * time.Second, MaxIdleConns: 20, MaxIdleConnsPerHost: 5, IdleConnTimeout: time.Minute, TLSHandshakeTimeout: 2 * time.Second, ResponseHeaderTimeout: time.Second}
	client, err := NewClient(config)
	if err != nil {
		t.Fatal(err)
	}
	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("Transport = %T", client.Transport)
	}
	if client.Timeout != config.Timeout || transport.MaxIdleConns != config.MaxIdleConns || transport.MaxIdleConnsPerHost != config.MaxIdleConnsPerHost || transport.IdleConnTimeout != config.IdleConnTimeout || transport.TLSHandshakeTimeout != config.TLSHandshakeTimeout || transport.ResponseHeaderTimeout != config.ResponseHeaderTimeout {
		t.Fatalf("client policy mismatch: %#v %#v", client, transport)
	}
	if transport == http.DefaultTransport {
		t.Fatal("mutated the shared default Transport")
	}
	if _, err := NewClient(Config{}); err == nil {
		t.Fatal("accepted empty policy")
	}
}
