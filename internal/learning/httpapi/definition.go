package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/MorseWayne/gogopher-arch/internal/learning/definition"
)

type DefinitionHandler struct{ registry *definition.Registry }

func NewDefinitionHandler(registry *definition.Registry) *DefinitionHandler {
	return &DefinitionHandler{registry: registry}
}

func (h *DefinitionHandler) Capability(w http.ResponseWriter, request *http.Request) {
	version, ok := requestedVersion(w, request)
	if !ok {
		return
	}
	releaseID := h.registry.CurrentReleaseID()
	value, err := h.registry.Get(definition.DefinitionRef{ReleaseID: releaseID, Kind: definition.KindCapability, ID: request.PathValue("id"), Version: version})
	if err != nil {
		h.writeDefinitionError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, struct {
		APIVersion  string          `json:"api_version"`
		ReleaseID   string          `json:"release_id"`
		ContentHash string          `json:"content_hash"`
		Capability  json.RawMessage `json:"capability"`
		Snapshot    any             `json:"snapshot"`
	}{APIVersion: APIVersion, ReleaseID: releaseID, ContentHash: value.ContentHash, Capability: value.Document, Snapshot: nil})
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
	writeJSON(w, http.StatusOK, struct {
		APIVersion string                  `json:"api_version"`
		ReleaseID  string                  `json:"release_id"`
		Activity   definition.ActivityView `json:"activity"`
	}{APIVersion: APIVersion, ReleaseID: releaseID, Activity: value})
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
