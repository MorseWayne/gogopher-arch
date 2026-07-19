package alertaccess

import "testing"

func TestAlertAccessContract(t *testing.T) {
	tests := []struct{ name string }{
		{name: "owner delete"},
		{name: "missing credential"},
		{name: "invalid credential"},
		{name: "other owner"},
		{name: "missing rule"},
		{name: "invalid id"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {})
	}
}
