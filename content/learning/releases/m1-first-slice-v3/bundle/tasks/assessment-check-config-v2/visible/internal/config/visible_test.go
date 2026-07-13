package config

import "testing"

func TestNormalizeSortsTargets(t *testing.T) {
	got, err := Normalize(Config{Targets: []Target{{Name: "z", URL: "https://z.example", TimeoutMS: 100}, {Name: "a", URL: "http://a.example", TimeoutMS: 50}}})
	if err != nil {
		t.Fatal(err)
	}
	if got.Targets[0].Name != "a" {
		t.Fatalf("first target = %q, want a", got.Targets[0].Name)
	}
}
