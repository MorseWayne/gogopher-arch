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
	request := httptest.NewRequest(http.MethodGet, "/api/v1/learning/capabilities/M1-01", nil)
	request.SetPathValue("id", "M1-01")
	request = authenticatedDefinitionRequest(request, "learner-1")
	response := httptest.NewRecorder()
	handler.Capability(response, request)
	if response.Code != http.StatusOK || reader.capabilityID != "M1-01" || reader.learnerID != "learner-1" || !reader.asOf.Equal(now) {
		t.Fatalf("response=%d reader=%#v body=%s", response.Code, reader, response.Body.String())
	}
	body := response.Body.String()
	for _, expected := range []string{
		`"version":3`, `"snapshot":null`, `"recent_evidence":[]`,
		`"definition":"release_bundle"`, `"snapshot":"capability_snapshots"`,
		`"evidence":"evidence_records"`, `"retention":"derived_at_read"`,
	} {
		if !strings.Contains(body, expected) {
			t.Fatalf("body missing %s: %s", expected, body)
		}
	}
}

func TestCapabilityCanReadAFrozenHistoricalReleaseAndVersion(t *testing.T) {
	registry := definitionTestRegistry(t)
	capability, err := registry.CapabilityView("m1-first-slice-v3", "M1-01", 1)
	if err != nil {
		t.Fatal(err)
	}
	reader := &learningReaderStub{capability: projection.CapabilityRead{
		ReleaseID: "m1-first-slice-v3", Capability: capability, RecentEvidence: []projection.EvidenceSummary{},
	}}
	handler, err := NewDefinitionHandler(registry, reader, DefinitionHandlerOptions{})
	if err != nil {
		t.Fatal(err)
	}
	request := authenticatedDefinitionRequest(httptest.NewRequest(
		http.MethodGet,
		"/api/v1/learning/capabilities/M1-01?version=1&release_id=m1-first-slice-v3",
		nil,
	), "learner-historical-capability")
	request.SetPathValue("id", "M1-01")
	response := httptest.NewRecorder()

	handler.Capability(response, request)

	if response.Code != http.StatusOK ||
		reader.selection != (projection.CapabilitySelection{ReleaseID: "m1-first-slice-v3", ID: "M1-01", Version: 1}) ||
		!strings.Contains(response.Body.String(), `"release_id":"m1-first-slice-v3"`) ||
		!strings.Contains(response.Body.String(), `"version":1`) {
		t.Fatalf("response=%d selection=%#v body=%s", response.Code, reader.selection, response.Body.String())
	}
}

func TestCapabilityRejectsInvalidExplicitVersion(t *testing.T) {
	registry := definitionTestRegistry(t)
	reader := &learningReaderStub{}
	handler, err := NewDefinitionHandler(registry, reader, DefinitionHandlerOptions{})
	if err != nil {
		t.Fatal(err)
	}
	request := authenticatedDefinitionRequest(httptest.NewRequest(
		http.MethodGet, "/api/v1/learning/capabilities/M1-01?version=0", nil,
	), "learner-invalid-capability")
	request.SetPathValue("id", "M1-01")
	response := httptest.NewRecorder()

	handler.Capability(response, request)

	if response.Code != http.StatusBadRequest ||
		!strings.Contains(response.Body.String(), `"code":"invalid_version"`) ||
		reader.capabilityID != "" {
		t.Fatalf("response=%d reader=%#v body=%s", response.Code, reader, response.Body.String())
	}
}

func TestActivityReturnsPublicTaskContextWithoutPrivateEvaluationRules(t *testing.T) {
	registry := definitionTestRegistry(t)
	handler, err := NewDefinitionHandler(registry, &learningReaderStub{}, DefinitionHandlerOptions{})
	if err != nil {
		t.Fatal(err)
	}
	request := authenticatedDefinitionRequest(httptest.NewRequest(http.MethodGet, "/api/v1/learning/activities/guided-run-model?version=7", nil), "learner-activity")
	request.SetPathValue("id", "guided-run-model")
	response := httptest.NewRecorder()
	handler.Activity(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("response=%d body=%s", response.Code, response.Body.String())
	}
	body := response.Body.String()
	for _, expected := range []string{`"task":`, `"allowed_actions":["build","submit","test","vet"]`, `"hints":[{"id":"return-a-string","level":1`, `"readme":"# 亲手完成第一个 Go 程序`, `"solution":"func welcome`} {
		if !strings.Contains(body, expected) {
			t.Fatalf("body missing %s: %s", expected, body)
		}
	}
	for _, forbidden := range []string{"held_out_tests", "race_tests", "assessment_rules", `"actions":`, "bundle_path", "welcome 的返回类型"} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("body contains private field %q: %s", forbidden, body)
		}
	}
}

func TestActivityCanReadAFrozenHistoricalRelease(t *testing.T) {
	registry := definitionTestRegistry(t)
	handler, err := NewDefinitionHandler(registry, &learningReaderStub{}, DefinitionHandlerOptions{})
	if err != nil {
		t.Fatal(err)
	}
	request := authenticatedDefinitionRequest(httptest.NewRequest(
		http.MethodGet,
		"/api/v1/learning/activities/guided-run-model?version=3&release_id=m1-first-slice-v3",
		nil,
	), "learner-historical-activity")
	request.SetPathValue("id", "guided-run-model")
	response := httptest.NewRecorder()

	handler.Activity(response, request)

	if response.Code != http.StatusOK ||
		!strings.Contains(response.Body.String(), `"release_id":"m1-first-slice-v3"`) ||
		!strings.Contains(response.Body.String(), `"version":3`) {
		t.Fatalf("response=%d body=%s", response.Code, response.Body.String())
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

func TestRoadmapReturnsServerDerivedCapabilityStates(t *testing.T) {
	registry := definitionTestRegistry(t)
	now := time.Date(2026, time.July, 19, 12, 0, 0, 0, time.UTC)
	capability, err := registry.CapabilityView(registry.CurrentReleaseID(), "M1-01", 3)
	if err != nil {
		t.Fatal(err)
	}
	reader := &learningReaderStub{roadmap: projection.RoadmapRead{
		ReleaseID: registry.CurrentReleaseID(),
		Items: []projection.RoadmapItem{{
			Capability: capability, Status: projection.RoadmapAvailable,
			HardPrerequisites: []projection.PrerequisiteStatus{}, RecommendedPrerequisites: []projection.PrerequisiteStatus{},
		}},
	}}
	handler, _ := NewDefinitionHandler(registry, reader, DefinitionHandlerOptions{Now: func() time.Time { return now }})
	request := authenticatedDefinitionRequest(httptest.NewRequest(http.MethodGet, "/api/v1/learning/roadmap", nil), "learner-roadmap")
	response := httptest.NewRecorder()

	handler.Roadmap(response, request)

	if response.Code != http.StatusOK || reader.learnerID != "learner-roadmap" || !reader.asOf.Equal(now) {
		t.Fatalf("response=%d reader=%#v body=%s", response.Code, reader, response.Body.String())
	}
	for _, expected := range []string{`"release_id":"m1-first-slice-v25"`, `"name":"编写并运行第一个 Go 程序"`, `"status":"available"`, `"state":"server_learning_state"`} {
		if !strings.Contains(response.Body.String(), expected) {
			t.Fatalf("body missing %s: %s", expected, response.Body.String())
		}
	}
}

type learningReaderStub struct {
	capability   projection.CapabilityRead
	next         *projection.NextRecommendation
	roadmap      projection.RoadmapRead
	err          error
	learnerID    string
	capabilityID string
	selection    projection.CapabilitySelection
	asOf         time.Time
}

func (s *learningReaderStub) Capability(_ context.Context, learnerID string, selection projection.CapabilitySelection, asOf time.Time) (projection.CapabilityRead, error) {
	s.learnerID, s.capabilityID, s.selection, s.asOf = learnerID, selection.ID, selection, asOf
	return s.capability, s.err
}

func (s *learningReaderStub) Next(_ context.Context, learnerID string, asOf time.Time) (*projection.NextRecommendation, error) {
	s.learnerID, s.asOf = learnerID, asOf
	return s.next, s.err
}

func (s *learningReaderStub) Roadmap(_ context.Context, learnerID string, asOf time.Time) (projection.RoadmapRead, error) {
	s.learnerID, s.asOf = learnerID, asOf
	return s.roadmap, s.err
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
