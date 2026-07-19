package alertdelivery

import "testing"

func TestManifestSmoke(t *testing.T) {
	entries := []Migration{{Name: "0001_create_alerts.up.sql", SHA256: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}}
	if err := ValidateManifest(entries); err != nil {
		t.Fatal(err)
	}
}
