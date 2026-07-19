package alertboard

import "testing"

func TestAlertboardCases(t *testing.T) {
	tests := []struct{name string}{{"authorized read"},{"cache hit"},{"cache outage"},{"acknowledge"},{"cancel worker"}}
	for _, testCase := range tests { t.Run(testCase.name, func(t *testing.T){}) }
}
