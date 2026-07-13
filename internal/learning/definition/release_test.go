package definition

import "testing"

func TestCanonicalJSONRFC8785GoldenVector(t *testing.T) {
	canonical, err := CanonicalJSON([]byte(`{"b":2,"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(canonical), `{"a":1,"b":2}`; got != want {
		t.Fatalf("CanonicalJSON() = %s, want %s", got, want)
	}
	if got, want := SHA256Hex(canonical), "43258cff783fe7036d8a43033f830adfc60ec037382473548ac742b888292777"; got != want {
		t.Fatalf("canonical hash = %s, want %s", got, want)
	}
}

func TestTaskBundleHashIsIndependentOfFormattingAndAssetOrder(t *testing.T) {
	assets := []AssetDigest{
		{Source: "starter/main.go", Path: "main.go", SHA256: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"},
		{Source: "starter/go.mod", Path: "go.mod", SHA256: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
	}

	first, err := TaskBundleHash([]byte(`{"version":1,"id":"task-v1"}`), assets)
	if err != nil {
		t.Fatal(err)
	}
	second, err := TaskBundleHash([]byte("{\n  \"id\": \"task-v1\",\n  \"version\": 1\n}\n"), []AssetDigest{assets[1], assets[0]})
	if err != nil {
		t.Fatal(err)
	}
	if first != second {
		t.Fatalf("hash differs by formatting/order: %s != %s", first, second)
	}
	if want := "46b92cb1f47889c8e55540afdd2ea647706aba7e269e1c1488c396247820cadd"; first != want {
		t.Fatalf("TaskBundleHash() = %s, want golden %s", first, want)
	}
}
