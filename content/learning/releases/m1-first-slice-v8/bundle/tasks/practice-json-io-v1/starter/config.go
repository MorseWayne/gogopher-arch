package checkcfg

import "fmt"

type Target struct {
	Name      string `json:"name"`
	URL       string `json:"url"`
	TimeoutMS int    `json:"timeout_ms"`
}

func Normalize(targets []Target) ([]Target, error) {
	if len(targets) == 0 {
		return nil, fmt.Errorf("targets must not be empty")
	}
	return targets, nil
}
