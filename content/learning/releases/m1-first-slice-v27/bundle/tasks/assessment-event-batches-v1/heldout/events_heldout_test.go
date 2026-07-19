package eventbatch

import "testing"

func TestSnapshotWindowOwnsStorage(t *testing.T) {
	source := []Event{{Key: "a", Value: 1}, {Key: "b", Value: 2}, {Key: "c", Value: 3}}
	window := SnapshotWindow(source, 1, 3)
	if len(window) != 2 {
		t.Fatalf("SnapshotWindow() = %#v", window)
	}
	source[1].Value = 20
	window[1].Value = 30
	if window[0].Value != 2 || source[2].Value != 3 {
		t.Fatalf("window and source still alias: source=%#v window=%#v", source, window)
	}
	if SnapshotWindow(source, -1, 1) != nil || SnapshotWindow(source, 2, 1) != nil || SnapshotWindow(source, 0, 4) != nil {
		t.Fatal("invalid window must return nil")
	}
}

func TestIndexLatestHandlesEmptyAndDuplicates(t *testing.T) {
	if empty := IndexLatest(nil); empty == nil || len(empty) != 0 {
		t.Fatalf("IndexLatest(nil) = %#v", empty)
	}
	got := IndexLatest([]Event{{Key: "x", Value: 1}, {Key: "x", Value: 2}, {Key: "y", Value: 4}})
	if len(got) != 2 || got["x"].Value != 2 || got["y"].Value != 4 {
		t.Fatalf("IndexLatest() = %#v", got)
	}
}
