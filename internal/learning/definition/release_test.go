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

func TestRuleSetHashIsIndependentOfFormattingAndCapabilityOrder(t *testing.T) {
	capabilities := []DefinitionDigest{
		{ID: "M1-03", Version: 1, ContentHash: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"},
		{ID: "M1-01", Version: 1, ContentHash: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
	}
	taskHash := "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"

	first, err := RuleSetHash([]byte(`{"version":1,"id":"activity"}`), capabilities, taskHash)
	if err != nil {
		t.Fatal(err)
	}
	second, err := RuleSetHash([]byte("{\n  \"id\": \"activity\",\n  \"version\": 1\n}"), []DefinitionDigest{capabilities[1], capabilities[0]}, taskHash)
	if err != nil {
		t.Fatal(err)
	}
	if first != second {
		t.Fatalf("hash differs by formatting/order: %s != %s", first, second)
	}
	if want := "473373b4cc01aa9baea5d11493e4f12e287ae5c962ebf158577d5257f48eba40"; first != want {
		t.Fatalf("RuleSetHash() = %s, want golden %s", first, want)
	}
}

func TestFullBundleHashIsIndependentOfFileOrder(t *testing.T) {
	files := []FileDigest{
		{Path: "tasks/example/main.go", SHA256: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"},
		{Path: "activities/example.json", SHA256: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
	}

	first, err := FullBundleHash(files)
	if err != nil {
		t.Fatal(err)
	}
	second, err := FullBundleHash([]FileDigest{files[1], files[0]})
	if err != nil {
		t.Fatal(err)
	}
	if first != second {
		t.Fatalf("hash differs by file order: %s != %s", first, second)
	}
	if want := "68d8ebd251302ad0a001b7194823fa4e6f5ada8a18f7330493d9d5525795bd12"; first != want {
		t.Fatalf("FullBundleHash() = %s, want golden %s", first, want)
	}
}
