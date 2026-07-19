package eventbatch

import "testing"

func TestSnapshotWindowAndIndexLatest(t *testing.T) {
	events := []Event{{Key: "api", Value: 1}, {Key: "db", Value: 2}, {Key: "api", Value: 3}}
	window := SnapshotWindow(events, 0, 2)
	if len(window) != 2 || window[1].Key != "db" {
		t.Fatalf("SnapshotWindow() = %#v", window)
	}
	index := IndexLatest(events)
	if len(index) != 2 || index["api"].Value != 3 || index["db"].Value != 2 {
		t.Fatalf("IndexLatest() = %#v", index)
	}
}
