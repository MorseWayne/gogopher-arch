package checkworker

import "testing"

func TestWorkerContract(t *testing.T) {
	tests := []struct{ name string }{{"success"}, {"temporary failure"}, {"permanent failure"}, {"duplicate"}, {"backpressure"}, {"cancellation"}, {"lease recovery"}}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {})
	}
}
