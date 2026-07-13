package httpapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/MorseWayne/gogopher-arch/internal/learning/review"
)

type ReviewClaimer interface {
	Claim(context.Context, string, string) (review.ClaimResult, error)
}

type ReviewHandler struct {
	claimer  ReviewClaimer
	observer AttemptObserver
}

func NewReviewHandler(claimer ReviewClaimer, observers ...AttemptObserver) (*ReviewHandler, error) {
	if claimer == nil {
		return nil, fmt.Errorf("review claimer is required")
	}
	handler := &ReviewHandler{claimer: claimer}
	if len(observers) > 0 {
		handler.observer = observers[0]
	}
	return handler, nil
}

func (h *ReviewHandler) Claim(w http.ResponseWriter, request *http.Request, reviewItemID string) {
	owner, ok := SessionFromContext(request.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthenticated", "learning session is required")
		return
	}
	result, err := h.claimer.Claim(request.Context(), owner.LearnerID, reviewItemID)
	if err != nil {
		switch {
		case errors.Is(err, review.ErrNotFound):
			writeError(w, http.StatusNotFound, "review_item_not_found", "review item not found")
		case errors.Is(err, review.ErrUnavailable):
			writeError(w, http.StatusConflict, "review_item_unavailable", "review item is no longer available")
		default:
			writeError(w, http.StatusInternalServerError, "review_unavailable", "review service is unavailable")
		}
		return
	}
	if result.Created && h.observer != nil {
		h.observer.AttemptCreated(result.Attempt)
	}
	status := http.StatusOK
	if result.Created {
		status = http.StatusCreated
	}
	writeJSON(w, status, attemptResponse(result.Attempt))
}
