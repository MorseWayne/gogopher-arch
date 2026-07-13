package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/MorseWayne/gogopher-arch/internal/learning/attempt"
	"github.com/MorseWayne/gogopher-arch/internal/learning/review"
	learningsession "github.com/MorseWayne/gogopher-arch/internal/learning/session"
)

func TestReviewHandlerClaimsForOwnerAndReplays(t *testing.T) {
	claimer := &reviewClaimerStub{result: review.ClaimResult{
		Attempt: attempt.Attempt{ID: "attempt-id", LearnerID: "owner", Mode: "review", Status: "active", Workspace: map[string]string{}},
		Created: true,
	}}
	handler, _ := NewReviewHandler(claimer)
	ctx := context.WithValue(context.Background(), sessionContextKey{}, learningsession.Session{LearnerID: "owner"})
	response := httptest.NewRecorder()
	handler.Claim(response, httptest.NewRequest(http.MethodPost, "/review-items/item-id/attempts", nil).WithContext(ctx), "item-id")
	if response.Code != http.StatusCreated || claimer.learnerID != "owner" || claimer.itemID != "item-id" || !strings.Contains(response.Body.String(), `"mode":"review"`) {
		t.Fatalf("claim response = %d %s input=%s/%s", response.Code, response.Body.String(), claimer.learnerID, claimer.itemID)
	}
	claimer.result.Created = false
	response = httptest.NewRecorder()
	handler.Claim(response, httptest.NewRequest(http.MethodPost, "/review-items/item-id/attempts", nil).WithContext(ctx), "item-id")
	if response.Code != http.StatusOK {
		t.Fatalf("replay status = %d", response.Code)
	}
}

func TestReviewHandlerUsesAsOfOnlyWhenEnabled(t *testing.T) {
	override := time.Date(2026, time.July, 20, 12, 0, 0, 0, time.UTC)
	claimer := &reviewClaimerStub{result: review.ClaimResult{Attempt: attempt.Attempt{Workspace: map[string]string{}}}}
	handler, _ := NewReviewHandlerWithOptions(claimer, ReviewHandlerOptions{AllowTestAsOf: true})
	ctx := context.WithValue(context.Background(), sessionContextKey{}, learningsession.Session{LearnerID: "owner"})
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/review-items/item-id/attempts?as_of="+override.Format(time.RFC3339Nano), nil).WithContext(ctx)
	handler.Claim(response, request, "item-id")
	if response.Code != http.StatusOK || claimer.timedClaims != 1 || !claimer.claimedAt.Equal(override) || claimer.claims != 0 {
		t.Fatalf("timed claim = status %d regular=%d timed=%d at=%s", response.Code, claimer.claims, claimer.timedClaims, claimer.claimedAt)
	}

	invalidResponse := httptest.NewRecorder()
	handler.Claim(invalidResponse, httptest.NewRequest(http.MethodPost, "/review-items/item-id/attempts?as_of=invalid", nil).WithContext(ctx), "item-id")
	if invalidResponse.Code != http.StatusBadRequest || claimer.timedClaims != 1 {
		t.Fatalf("invalid override = status %d timed=%d", invalidResponse.Code, claimer.timedClaims)
	}

	localHandler, _ := NewReviewHandler(claimer)
	localResponse := httptest.NewRecorder()
	localHandler.Claim(localResponse, httptest.NewRequest(http.MethodPost, "/review-items/item-id/attempts?as_of=invalid", nil).WithContext(ctx), "item-id")
	if localResponse.Code != http.StatusOK || claimer.claims != 1 {
		t.Fatalf("local override = status %d regular=%d", localResponse.Code, claimer.claims)
	}
}

func TestReviewHandlerHidesOtherOwnersAndRejectsInactiveItems(t *testing.T) {
	claimer := &reviewClaimerStub{err: review.ErrNotFound}
	handler, _ := NewReviewHandler(claimer)
	ctx := context.WithValue(context.Background(), sessionContextKey{}, learningsession.Session{LearnerID: "owner"})
	response := httptest.NewRecorder()
	handler.Claim(response, httptest.NewRequest(http.MethodPost, "/review-items/item-id/attempts", nil).WithContext(ctx), "item-id")
	if response.Code != http.StatusNotFound {
		t.Fatalf("not found status = %d", response.Code)
	}
	claimer.err = review.ErrUnavailable
	response = httptest.NewRecorder()
	handler.Claim(response, httptest.NewRequest(http.MethodPost, "/review-items/item-id/attempts", nil).WithContext(ctx), "item-id")
	if response.Code != http.StatusConflict {
		t.Fatalf("unavailable status = %d", response.Code)
	}
}

type reviewClaimerStub struct {
	result      review.ClaimResult
	err         error
	learnerID   string
	itemID      string
	claimedAt   time.Time
	claims      int
	timedClaims int
}

func (s *reviewClaimerStub) Claim(_ context.Context, learnerID, itemID string) (review.ClaimResult, error) {
	s.learnerID, s.itemID = learnerID, itemID
	s.claims++
	return s.result, s.err
}

func (s *reviewClaimerStub) ClaimAt(_ context.Context, learnerID, itemID string, claimedAt time.Time) (review.ClaimResult, error) {
	s.learnerID, s.itemID, s.claimedAt = learnerID, itemID, claimedAt
	s.timedClaims++
	return s.result, s.err
}
