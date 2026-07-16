package checkcfg

import (
	"fmt"
	"strings"
)

func NormalizeName(value string) (string, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return "", fmt.Errorf("name must not be empty")
	}
	return value, nil
}
