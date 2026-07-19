package checkcfg

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	Name string `json:"name"`
}

func Load(path string) (Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return Config{}, fmt.Errorf("open config: %v", err)
	}
	defer f.Close()

	var cfg Config
	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		return Config{}, fmt.Errorf("decode config: %v", err)
	}
	return cfg, nil
}
