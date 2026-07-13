package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
	result    review.ClaimResult
	err       error
	learnerID string
	itemID    string
}

func (s *reviewClaimerStub) Claim(_ context.Context, learnerID, itemID string) (review.ClaimResult, error) {
	s.learnerID, s.itemID = learnerID, itemID
	return s.result, s.err
}
