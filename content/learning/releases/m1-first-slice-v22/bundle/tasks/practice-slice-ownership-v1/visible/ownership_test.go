package ownership

import "testing"

func TestCloneWindowDetachesAndCountTagsInitializes(t *testing.T) {
	source := []byte("gopher")
	got := CloneWindow(source, 1, 4)
	source[2] = 'X'
	if string(got) != "oph" {
		t.Fatalf("CloneWindow() = %q after source mutation", got)
	}
	counts := CountTags([]string{"api", "db", "api"})
	if counts == nil || counts["api"] != 2 || counts["db"] != 1 {
		t.Fatalf("CountTags() = %#v", counts)
	}
	if empty := CountTags(nil); empty == nil || len(empty) != 0 {
		t.Fatalf("CountTags(nil) = %#v, want non-nil empty map", empty)
	}
}
