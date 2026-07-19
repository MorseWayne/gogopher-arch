package checkcfg

import "testing"

func TestNormalizeNameContract(t *testing.T) {
	if _, err := NormalizeName("   "); err == nil {
		t.Fatal("NormalizeName(empty) error = nil")
	}
}
