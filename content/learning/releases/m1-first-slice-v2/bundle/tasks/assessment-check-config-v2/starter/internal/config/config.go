package config

import "fmt"

type Config struct {
	Targets []Target `json:"targets"`
}

type Target struct {
	Name      string `json:"name"`
	URL       string `json:"url"`
	TimeoutMS int    `json:"timeout_ms"`
}

func Load(path string) (Config, error) {
	return Config{}, fmt.Errorf("Load %q: not implemented", path)
}

func Normalize(cfg Config) (Config, error) {
	return cfg, nil
}
