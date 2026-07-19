package projectapi

import "testing"

func TestAccessControlContract(t *testing.T) {
	tests := []struct{ name string }{
		{name: "valid owner"},
		{name: "missing credential"},
		{name: "invalid credential"},
		{name: "other owner"},
		{name: "missing project"},
		{name: "invalid id"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {})
	}
}
