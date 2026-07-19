package trace

import (
	"reflect"
	"testing"
)

func TestMinifluxCallChainsMatchFixedCommit(t *testing.T) {
	wantCategory := []Step{
		{Path: "internal/api/api.go", Symbol: "POST /v1/categories", Responsibility: "route"},
		{Path: "internal/api/category_handlers.go", Symbol: "createCategoryHandler", Responsibility: "decode-and-respond"},
		{Path: "internal/validator/category.go", Symbol: "ValidateCategoryCreation", Responsibility: "validate"},
		{Path: "internal/storage/category.go", Symbol: "CreateCategory", Responsibility: "persist"},
	}
	wantRefresh := []Step{
		{Path: "internal/cli/scheduler.go", Symbol: "feedScheduler", Responsibility: "schedule"},
		{Path: "internal/storage/batch.go", Symbol: "FetchJobs", Responsibility: "select-due-feeds"},
		{Path: "internal/worker/pool.go", Symbol: "Push", Responsibility: "backpressure"},
		{Path: "internal/worker/worker.go", Symbol: "Run", Responsibility: "consume-job"},
		{Path: "internal/reader/handler/handler.go", Symbol: "RefreshFeed", Responsibility: "refresh-and-store"},
	}
	if got := CategoryCreation(); !reflect.DeepEqual(got, wantCategory) {
		t.Fatalf("CategoryCreation() = %#v", got)
	}
	if got := FeedRefresh(); !reflect.DeepEqual(got, wantRefresh) {
		t.Fatalf("FeedRefresh() = %#v", got)
	}
}
