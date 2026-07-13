package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/MorseWayne/gogopher-arch/internal/learning/definition"
	"github.com/MorseWayne/gogopher-arch/internal/learning/projection"
	learningsession "github.com/MorseWayne/gogopher-arch/internal/learning/session"
)

func TestCapabilityReturnsCurrentDefinitionAndExplicitStateSources(t *testing.T) {
	registry := definitionTestRegistry(t)
	stored, _ := registry.Latest(registry.CurrentReleaseID(), definition.KindCapability, "M1-01")
	capability, _ := registry.CapabilityView(registry.CurrentReleaseID(), stored.ID, stored.Version)
	now := time.Date(2026, time.July, 13, 12, 0, 0, 0, time.UTC)
	reader := &learningReaderStub{capability: projection.CapabilityRead{
		ReleaseID: registry.CurrentReleaseID(), Capability: capability, RecentEvidence: []projection.EvidenceSummary{},
	}}
	handler, err := NewDefinitionHandler(registry, reader, DefinitionHandlerOptions{Now: func() time.Time { return now }})
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodGet, "/api/v1/learning/capabilities/M1-01?version=1", nil)
	request.SetPathValue("id", "M1-01")
	request = authenticatedDefinitionRequest(request, "learner-1")
	response := httptest.NewRecorder()
	handler.Capability(response, request)
	if response.Code != http.StatusOK || reader.capabilityID != "M1-01" || reader.learnerID != "learner-1" || !reader.asOf.Equal(now) {
		t.Fatalf("response=%d reader=%#v body=%s", response.Code, reader, response.Body.String())
	}
	body := response.Body.String()
	for _, expected := range []string{
		`"version":2`, `"snapshot":null`, `"recent_evidence":[]`,
		`"definition":"release_bundle"`, `"snapshot":"capability_snapshots"`,
		`"evidence":"evidence_records"`, `"retention":"derived_at_read"`,
	} {
		if !strings.Contains(body, expected) {
			t.Fatalf("body missing %s: %s", expected, body)
		}
	}
}

func TestActivityReturnsPublicTaskContextWithoutPrivateEvaluationRules(t *testing.T) {
	registry := definitionTestRegistry(t)
	handler, err := NewDefinitionHandler(registry, &learningReaderStub{}, DefinitionHandlerOptions{})
	if err != nil {
		t.Fatal(err)
	}
	request := authenticatedDefinitionRequest(httptest.NewRequest(http.MethodGet, "/api/v1/learning/activities/guided-run-model?version=3", nil), "learner-activity")
	request.SetPathValue("id", "guided-run-model")
	response := httptest.NewRecorder()
	handler.Activity(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("response=%d body=%s", response.Code, response.Body.String())
	}
	body := response.Body.String()
	for _, expected := range []string{`"task":`, `"allowed_actions":["build","test","vet"]`, `"hints":[{"id":"read-first-error","level":1`, `"readme":"# 读懂工具链反馈`} {
		if !strings.Contains(body, expected) {
			t.Fatalf("body missing %s: %s", expected, body)
		}
	}
	for _, forbidden := range []string{"held_out_tests", "assessment_rules", `"actions":`, "bundle_path", "从输出顶部找到"} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("body contains private field %q: %s", forbidden, body)
		}
	}
}

func TestNextUsesTestOverrideAndReturnsNullWhenNoActivityExists(t *testing.T) {
	registry := definitionTestRegistry(t)
	serverNow := time.Date(2026, time.July, 13, 12, 0, 0, 0, time.UTC)
	override := serverNow.Add(72 * time.Hour)
	reader := &learningReaderStub{}
	handler, _ := NewDefinitionHandler(registry, reader, DefinitionHandlerOptions{
		Now: func() time.Time { return serverNow }, AllowTestAsOf: true,
	})
	request := httptest.NewRequest(http.MethodGet, "/api/v1/learning/next?as_of="+url.QueryEscape(override.Format(time.RFC3339Nano)), nil)
	request = authenticatedDefinitionRequest(request, "learner-2")
	response := httptest.NewRecorder()
	handler.Next(response, request)
	if response.Code != http.StatusOK || !reader.asOf.Equal(override) {
		t.Fatalf("response=%d as_of=%s body=%s", response.Code, reader.asOf, response.Body.String())
	}
	var body struct {
		Recommendation *projection.NextRecommendation `json:"recommendation"`
		Source         struct {
			Clock string `json:"clock"`
		} `json:"source"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Recommendation != nil || body.Source.Clock != "test_override" {
		t.Fatalf("body = %#v", body)
	}
}

func TestNextIgnoresClientAsOfOutsideTestEnvironment(t *testing.T) {
	registry := definitionTestRegistry(t)
	serverNow := time.Date(2026, time.July, 13, 12, 0, 0, 0, time.UTC)
	reader := &learningReaderStub{}
	handler, _ := NewDefinitionHandler(registry, reader, DefinitionHandlerOptions{Now: func() time.Time { return serverNow }})
	request := httptest.NewRequest(http.MethodGet, "/api/v1/learning/next?as_of=not-a-time", nil)
	request = authenticatedDefinitionRequest(request, "learner-3")
	response := httptest.NewRecorder()
	handler.Next(response, request)
	if response.Code != http.StatusOK || !reader.asOf.Equal(serverNow) || !strings.Contains(response.Body.String(), `"clock":"server"`) {
		t.Fatalf("response=%d as_of=%s body=%s", response.Code, reader.asOf, response.Body.String())
	}
}

func TestNextRejectsInvalidTestAsOf(t *testing.T) {
	registry := definitionTestRegistry(t)
	reader := &learningReaderStub{}
	handler, _ := NewDefinitionHandler(registry, reader, DefinitionHandlerOptions{AllowTestAsOf: true})
	request := authenticatedDefinitionRequest(httptest.NewRequest(http.MethodGet, "/api/v1/learning/next?as_of=invalid", nil), "learner-4")
	response := httptest.NewRecorder()
	handler.Next(response, request)
	if response.Code != http.StatusBadRequest || !strings.Contains(response.Body.String(), `"code":"invalid_as_of"`) {
		t.Fatalf("response=%d body=%s", response.Code, response.Body.String())
	}
}

type learningReaderStub struct {
	capability   projection.CapabilityRead
	next         *projection.NextRecommendation
	err          error
	learnerID    string
	capabilityID string
	asOf         time.Time
}

func (s *learningReaderStub) Capability(_ context.Context, learnerID, capabilityID string, asOf time.Time) (projection.CapabilityRead, error) {
	s.learnerID, s.capabilityID, s.asOf = learnerID, capabilityID, asOf
	return s.capability, s.err
}

func (s *learningReaderStub) Next(_ context.Context, learnerID string, asOf time.Time) (*projection.NextRecommendation, error) {
	s.learnerID, s.asOf = learnerID, asOf
	return s.next, s.err
}

func definitionTestRegistry(t *testing.T) *definition.Registry {
	t.Helper()
	contentDir, err := filepath.Abs("../../../content/learning")
	if err != nil {
		t.Fatal(err)
	}
	registry, err := definition.LoadRegistry(definition.RegistryOptions{ContentDir: contentDir})
	if err != nil {
		t.Fatal(err)
	}
	return registry
}

func authenticatedDefinitionRequest(request *http.Request, learnerID string) *http.Request {
	return request.WithContext(context.WithValue(request.Context(), sessionContextKey{}, learningsession.Session{LearnerID: learnerID}))
}
