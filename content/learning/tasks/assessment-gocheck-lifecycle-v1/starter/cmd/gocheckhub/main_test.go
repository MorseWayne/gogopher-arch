package main

import "testing"

func TestLifecycleContract(t *testing.T) {
	tests := []struct{ name string }{{name: "invalid config"}, {name: "open failure"}, {name: "handler failure"}, {name: "server failure"}, {name: "cancellation"}, {name: "serve failure"}}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {})
	}
}
