package httpapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/MorseWayne/gogopher-arch/internal/learning/definition"
	"github.com/MorseWayne/gogopher-arch/internal/learning/projection"
)

type LearningReader interface {
	Capability(context.Context, string, string, time.Time) (projection.CapabilityRead, error)
	Next(context.Context, string, time.Time) (*projection.NextRecommendation, error)
}

type DefinitionHandlerOptions struct {
	Now           func() time.Time
	AllowTestAsOf bool
}

type DefinitionHandler struct {
	registry      *definition.Registry
	reader        LearningReader
	now           func() time.Time
	allowTestAsOf bool
}

func NewDefinitionHandler(registry *definition.Registry, reader LearningReader, options DefinitionHandlerOptions) (*DefinitionHandler, error) {
	if registry == nil || reader == nil {
		return nil, fmt.Errorf("definition registry and learning reader are required")
	}
	if options.Now == nil {
		options.Now = time.Now
	}
	return &DefinitionHandler{registry: registry, reader: reader, now: options.Now, allowTestAsOf: options.AllowTestAsOf}, nil
}

func (h *DefinitionHandler) Capability(w http.ResponseWriter, request *http.Request) {
	owner, ok := SessionFromContext(request.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthenticated", "learning session is required")
		return
	}
	asOf := h.now().UTC()
	value, err := h.reader.Capability(request.Context(), owner.LearnerID, request.PathValue("id"), asOf)
	if err != nil {
		h.writeDefinitionError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, struct {
		APIVersion string `json:"api_version"`
		projection.CapabilityRead
		Source capabilitySource `json:"source"`
	}{
		APIVersion: APIVersion, CapabilityRead: value,
		Source: capabilitySource{
			Definition: "release_bundle", Snapshot: "capability_snapshots", Evidence: "evidence_records",
			Retention: "derived_at_read", AsOf: asOf, Clock: "server",
		},
	})
}

func (h *DefinitionHandler) Activity(w http.ResponseWriter, request *http.Request) {
	version, ok := requestedVersion(w, request)
	if !ok {
		return
	}
	releaseID := h.registry.CurrentReleaseID()
	value, err := h.registry.ActivityView(releaseID, request.PathValue("id"), version)
	if err != nil {
		h.writeDefinitionError(w, err)
		return
	}
	task, err := h.registry.TaskView(releaseID, value.TaskRef.ID, value.TaskRef.Version)
	if err != nil {
		h.writeDefinitionError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, struct {
		APIVersion string                  `json:"api_version"`
		ReleaseID  string                  `json:"release_id"`
		Activity   definition.ActivityView `json:"activity"`
		Task       definition.TaskView     `json:"task"`
	}{APIVersion: APIVersion, ReleaseID: releaseID, Activity: value, Task: task})
}

func (h *DefinitionHandler) Next(w http.ResponseWriter, request *http.Request) {
	owner, ok := SessionFromContext(request.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthenticated", "learning session is required")
		return
	}
	asOf, clock, ok := h.asOf(w, request)
	if !ok {
		return
	}
	value, err := h.reader.Next(request.Context(), owner.LearnerID, asOf)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "learning_next_unavailable", "next learning activity is unavailable")
		return
	}
	writeJSON(w, http.StatusOK, struct {
		APIVersion     string                         `json:"api_version"`
		Recommendation *projection.NextRecommendation `json:"recommendation"`
		Source         nextSource                     `json:"source"`
	}{
		APIVersion: APIVersion, Recommendation: value,
		Source: nextSource{ReleaseID: h.registry.CurrentReleaseID(), State: "server_learning_state", AsOf: asOf, Clock: clock},
	})
}

type capabilitySource struct {
	Definition string    `json:"definition"`
	Snapshot   string    `json:"snapshot"`
	Evidence   string    `json:"evidence"`
	Retention  string    `json:"retention"`
	AsOf       time.Time `json:"as_of"`
	Clock      string    `json:"clock"`
}

type nextSource struct {
	ReleaseID string    `json:"release_id"`
	State     string    `json:"state"`
	AsOf      time.Time `json:"as_of"`
	Clock     string    `json:"clock"`
}

func (h *DefinitionHandler) asOf(w http.ResponseWriter, request *http.Request) (time.Time, string, bool) {
	now := h.now().UTC()
	raw := request.URL.Query().Get("as_of")
	if !h.allowTestAsOf || raw == "" {
		return now, "server", true
	}
	value, err := time.Parse(time.RFC3339Nano, raw)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_as_of", "as_of must be an RFC3339 timestamp")
		return time.Time{}, "", false
	}
	return value.UTC(), "test_override", true
}

func requestedVersion(w http.ResponseWriter, request *http.Request) (int, bool) {
	raw := request.URL.Query().Get("version")
	if raw == "" {
		return 1, true
	}
	version, err := strconv.Atoi(raw)
	if err != nil || version < 1 {
		writeError(w, http.StatusBadRequest, "invalid_version", "version must be a positive integer")
		return 0, false
	}
	return version, true
}
func (h *DefinitionHandler) writeDefinitionError(w http.ResponseWriter, err error) {
	if errors.Is(err, definition.ErrDefinitionNotFound) {
		writeError(w, http.StatusNotFound, "definition_not_found", "definition not found")
		return
	}
	writeError(w, http.StatusInternalServerError, "definition_unavailable", "definition is unavailable")
}
