package trace

import "testing"

func TestTraceStartsAtRealEntrypoints(t *testing.T) {
	category := CategoryCreation()
	if len(category) == 0 || category[0].Path != "internal/api/api.go" || category[0].Symbol != "POST /v1/categories" {
		t.Fatalf("category trace starts at %#v", category)
	}
	refresh := FeedRefresh()
	if len(refresh) == 0 || refresh[0].Path != "internal/cli/scheduler.go" || refresh[0].Symbol != "feedScheduler" {
		t.Fatalf("refresh trace starts at %#v", refresh)
	}
}
