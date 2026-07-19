package checkcfg

import "testing"

func TestNormalizeSortsAndValidates(t *testing.T) {
	got, err := Normalize([]Target{{Name: "z", URL: "https://z.example", TimeoutMS: 1}, {Name: "a", URL: "http://a.example", TimeoutMS: 1}})
	if err != nil {
		t.Fatal(err)
	}
	if got[0].Name != "a" {
		t.Fatalf("first target = %q, want a", got[0].Name)
	}
	if _, err := Normalize([]Target{{Name: "bad", URL: "ftp://example", TimeoutMS: 0}}); err == nil {
		t.Fatal("Normalize() error = nil, want validation error")
	}
}
