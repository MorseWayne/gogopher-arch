package merge

import "testing"

func TestMergeOverridesByName(t *testing.T) {
	got, err := Merge(Document{Services: []Service{{Name: "api", Endpoint: "https://old.example", Retry: 1}}}, Document{Services: []Service{{Name: "api", Endpoint: "https://new.example", Retry: 2}}})
	if err != nil {
		t.Fatal(err)
	}
	if got.Services[0].Endpoint != "https://new.example" {
		t.Fatalf("endpoint = %q", got.Services[0].Endpoint)
	}
}
