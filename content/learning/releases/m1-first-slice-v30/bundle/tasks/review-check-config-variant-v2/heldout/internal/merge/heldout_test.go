package merge

import "testing"

func TestMergeRejectsDuplicateAndInvalidRetry(t *testing.T) {
	_, err := Merge(Document{Services: []Service{{Name: "api", Endpoint: "https://api.example", Retry: -1}}}, Document{})
	if err == nil {
		t.Fatal("Merge(invalid retry) error = nil")
	}
}
