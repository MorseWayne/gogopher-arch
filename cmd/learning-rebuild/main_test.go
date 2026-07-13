package main

import (
	"io"
	"testing"
)

func TestParseOptionsDefaultsToDryRunAndRequiresExplicitApply(t *testing.T) {
	options, err := parseOptions([]string{"--learner-id", "learner", "--capability-id", "M1-03", "--dry-run"}, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	if options.Apply || options.LearnerID != "learner" || options.CapabilityID != "M1-03" || options.AsOf.IsZero() {
		t.Fatalf("options = %#v", options)
	}
	options, err = parseOptions([]string{"--apply"}, io.Discard)
	if err != nil || !options.Apply {
		t.Fatalf("apply options = %#v, %v", options, err)
	}
	if _, err := parseOptions([]string{"--dry-run", "--apply"}, io.Discard); err == nil {
		t.Fatal("conflicting modes error = nil")
	}
}
