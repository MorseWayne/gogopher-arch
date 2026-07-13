package config

import (
	"fmt"
	"net"
)

type SandboxConfig struct {
	ListenAddress string
}

func LoadSandbox() (SandboxConfig, error) {
	config := SandboxConfig{ListenAddress: env("SANDBOX_LISTEN_ADDRESS", "127.0.0.1:8081")}
	if _, _, err := net.SplitHostPort(config.ListenAddress); err != nil {
		return SandboxConfig{}, fmt.Errorf("invalid SANDBOX_LISTEN_ADDRESS: %w", err)
	}
	return config, nil
}
