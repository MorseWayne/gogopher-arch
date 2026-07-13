package projection

import (
	"context"
	"errors"
	"testing"
	"time"
	"unicode/utf8"
)

func TestWorkerCompletesVersionedProjectionRequest(t *testing.T) {
	now := time.Date(2026, time.July, 13, 13, 0, 0, 0, time.UTC)
	repository := &projectionRequestRepositoryStub{request: Request{ID: "request-1", AttemptCount: 1}}
	projector := &requestProjectorStub{}
	worker, err := NewWorker(repository, projector, projectionWorkerTestOptions(now))
	if err != nil {
		t.Fatal(err)
	}
	processed, err := worker.RunOnce(t.Context())
	if err != nil || !processed {
		t.Fatalf("RunOnce() = %v, %v", processed, err)
	}
	if projector.calls != 1 || repository.completedConsumer != projectionConsumer ||
		repository.completedVersion != ProjectionConsumerVersion {
		t.Fatalf("projector calls=%d consumer=%q version=%d",
			projector.calls, repository.completedConsumer, repository.completedVersion)
	}
}

func TestWorkerRetriesWithBoundedBackoffAndRecordsExhaustion(t *testing.T) {
	now := time.Date(2026, time.July, 13, 13, 0, 0, 0, time.UTC)
	observer := &projectionObserverStub{}
	repository := &projectionRequestRepositoryStub{
		request:     Request{ID: "request-2", AttemptCount: 3},
		retryResult: RetryResult{AttemptCount: 3, Exhausted: true},
	}
	options := projectionWorkerTestOptions(now)
	options.MaxAttempts = 3
	options.MaxBackoff = 3 * time.Second
	options.Observer = observer
	worker, err := NewWorker(repository, &requestProjectorStub{err: errors.New("projection unavailable\ninternal detail")}, options)
	if err != nil {
		t.Fatal(err)
	}
	processed, err := worker.RunOnce(t.Context())
	if err != nil || !processed {
		t.Fatalf("RunOnce() = %v, %v", processed, err)
	}
	if repository.retryDelay != 3*time.Second || repository.retrySummary != "projection unavailable\ninternal detail" {
		t.Fatalf("retry delay=%s summary=%q", repository.retryDelay, repository.retrySummary)
	}
	if observer.calls != 1 || !observer.exhausted {
		t.Fatalf("observer = calls %d exhausted %v", observer.calls, observer.exhausted)
	}
}

func TestSummarizeErrorIsBoundedAndSingleLine(t *testing.T) {
	value := summarizeError("  first\n\tsecond  ")
	if value != "first second" {
		t.Fatalf("summarizeError() = %q", value)
	}
	long := summarizeError(string(make([]byte, 600)))
	if len(long) > 512 {
		t.Fatalf("summarizeError() length = %d", len(long))
	}
	multibyte := summarizeError(string(make([]byte, 510)) + "错误")
	if len(multibyte) > 512 || !utf8.ValidString(multibyte) {
		t.Fatalf("multibyte summarizeError() = length %d valid %v", len(multibyte), utf8.ValidString(multibyte))
	}
}

type projectionRequestRepositoryStub struct {
	request           Request
	retryResult       RetryResult
	completedConsumer string
	completedVersion  int
	retryDelay        time.Duration
	retrySummary      string
}

func (s *projectionRequestRepositoryStub) ClaimRequest(context.Context, string, time.Time, time.Duration) (Request, bool, error) {
	return s.request, true, nil
}

func (s *projectionRequestRepositoryStub) CompleteRequest(_ context.Context, _, _, consumer string, version int, _ time.Time) error {
	s.completedConsumer = consumer
	s.completedVersion = version
	return nil
}

func (s *projectionRequestRepositoryStub) RetryRequest(_ context.Context, _, _ string, _ time.Time, delay time.Duration, _ int, summary string) (RetryResult, error) {
	s.retryDelay = delay
	s.retrySummary = summary
	return s.retryResult, nil
}

type requestProjectorStub struct {
	calls int
	err   error
}

func (s *requestProjectorStub) RebuildRequest(context.Context, Request, time.Time) error {
	s.calls++
	return s.err
}

type projectionObserverStub struct {
	calls     int
	exhausted bool
}

func (s *projectionObserverStub) ProjectionRetried(exhausted bool) {
	s.calls++
	s.exhausted = exhausted
}

func projectionWorkerTestOptions(now time.Time) WorkerOptions {
	return WorkerOptions{
		Owner: "projection-worker", Lease: 10 * time.Second, PollInterval: time.Millisecond,
		MaxAttempts: 3, BaseBackoff: time.Second, MaxBackoff: time.Minute,
		Now: func() time.Time { return now },
	}
}
