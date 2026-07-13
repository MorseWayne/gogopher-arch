package config

import "testing"

func TestLoadSandboxDefaultsToLoopback(t *testing.T) {
	t.Setenv("SANDBOX_LISTEN_ADDRESS", "")
	config, err := LoadSandbox()
	if err != nil {
		t.Fatal(err)
	}
	if config.ListenAddress != "127.0.0.1:8081" {
		t.Fatalf("listen address = %q", config.ListenAddress)
	}
}
