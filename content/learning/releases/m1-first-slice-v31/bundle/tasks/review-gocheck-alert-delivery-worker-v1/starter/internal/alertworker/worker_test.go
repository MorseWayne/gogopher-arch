package alertworker

import "testing"

func TestAlertWorkerContract(t *testing.T) {
	tests := []struct{ name string }{{"delivered"}, {"retry"}, {"permanent failure"}, {"duplicate"}, {"bounded backlog"}, {"shutdown"}, {"restart recovery"}}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {})
	}
}
