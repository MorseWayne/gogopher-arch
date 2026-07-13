package projection

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/MorseWayne/gogopher-arch/internal/learning/definition"
	platformdb "github.com/MorseWayne/gogopher-arch/internal/platform/database"
)

func TestSelectAcquisitionHonorsCurrentVersionAndPrerequisites(t *testing.T) {
	capabilities := []definition.CapabilityView{
		{ID: "M1-01", Version: 2, ContentHash: hash64("a")},
		{
			ID: "M1-09", Version: 1, ContentHash: hash64("b"),
			Prerequisites: definition.CapabilityPrerequisitesView{
				Hard:        []definition.VersionedDefinitionRef{{ID: "M1-01", Version: 1}},
				Recommended: []definition.VersionedDefinitionRef{{ID: "M1-03", Version: 1}},
			},
		},
	}
	activities := []definition.ActivityView{
		{ID: "guided", Version: 2, Mode: "guided", CapabilityRefs: []definition.VersionedDefinitionRef{{ID: "M1-01", Version: 2}}},
		{ID: "practice-tests", Version: 1, Mode: "practice", CapabilityRefs: []definition.VersionedDefinitionRef{{ID: "M1-09", Version: 1}}},
	}

	oldVersionOnly := map[string]currentState{
		"M1-01": {Version: 1, Hash: hash64("c"), Acquisition: AcquisitionStable},
	}
	first := selectAcquisition(capabilities, activities, oldVersionOnly)
	if first == nil || first.TargetCapability == nil || *first.TargetCapability != (definition.VersionedDefinitionRef{ID: "M1-01", Version: 2}) {
		t.Fatalf("old version recommendation = %#v", first)
	}

	currentRoot := map[string]currentState{
		"M1-01": {Version: 2, Hash: hash64("a"), Acquisition: AcquisitionVerified},
	}
	next := selectAcquisition(capabilities, activities, currentRoot)
	if next == nil || next.Activity.ID != "practice-tests" || len(next.HardPrerequisites) != 1 || !next.HardPrerequisites[0].Satisfied || next.HardPrerequisites[0].SatisfiedVersion != 2 {
		t.Fatalf("unblocked recommendation = %#v", next)
	}
	if len(next.RecommendedPrerequisites) != 1 || next.RecommendedPrerequisites[0].Satisfied {
		t.Fatalf("recommended prerequisite must explain without blocking: %#v", next.RecommendedPrerequisites)
	}
}

func TestSelectAcquisitionReturnsNilForBlockedOrUnavailablePath(t *testing.T) {
	capability := definition.CapabilityView{
		ID: "M1-07", Version: 1, ContentHash: hash64("a"),
		Prerequisites: definition.CapabilityPrerequisitesView{
			Hard: []definition.VersionedDefinitionRef{{ID: "M1-03", Version: 1}},
		},
	}
	activity := definition.ActivityView{
		ID: "practice-json", Version: 1, Mode: "practice",
		CapabilityRefs: []definition.VersionedDefinitionRef{{ID: "M1-07", Version: 1}},
	}
	if value := selectAcquisition([]definition.CapabilityView{capability}, []definition.ActivityView{activity}, nil); value != nil {
		t.Fatalf("blocked recommendation = %#v", value)
	}
	if value := selectAcquisition([]definition.CapabilityView{{ID: "M1-01", Version: 2, ContentHash: hash64("b")}}, nil, nil); value != nil {
		t.Fatalf("unavailable recommendation = %#v", value)
	}
}

func TestReaderDoesNotInheritPreviousCapabilityVersion(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set TEST_DATABASE_URL to run PostgreSQL integration tests")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	db, err := platformdb.Open(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	schema := fmt.Sprintf("reader_test_%d", time.Now().UnixNano())
	if _, err := db.ExecContext(ctx, `CREATE SCHEMA "`+schema+`"`); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_, _ = db.ExecContext(context.Background(), `DROP SCHEMA "`+schema+`" CASCADE`)
		_ = db.Close()
	}()
	migrator, _ := platformdb.NewMigrator(db, os.DirFS("../../../db/migrations"), platformdb.MigratorOptions{Schema: schema})
	if err := migrator.Up(ctx); err != nil {
		t.Fatal(err)
	}
	contentDir, _ := filepath.Abs("../../../content/learning")
	registry, err := definition.LoadRegistry(definition.RegistryOptions{ContentDir: contentDir})
	if err != nil {
		t.Fatal(err)
	}
	history, _ := definition.NewReleaseStore(db, definition.ReleaseStoreOptions{Schema: schema})
	if err := history.Register(ctx, registry); err != nil {
		t.Fatal(err)
	}
	learnerID := "00000000-0000-4000-8000-000000009001"
	now := time.Date(2026, time.July, 13, 12, 0, 0, 0, time.UTC)
	oldCapability, _ := registry.CapabilityView(registry.CurrentReleaseID(), "M1-01", 1)
	if _, err := db.ExecContext(ctx, `INSERT INTO "`+schema+`".learners (id,created_at) VALUES ($1,$2)`, learnerID, now); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `
		INSERT INTO "`+schema+`".capability_snapshots (
			learner_id,capability_id,capability_version,capability_hash,
			acquisition_state,independence_state,transfer_state,retention_base_state,
			projection_version,projected_at
		) VALUES ($1,'M1-01',1,$3,'stable','independent','variant','fresh',1,$2)`,
		learnerID, now, oldCapability.ContentHash); err != nil {
		t.Fatal(err)
	}
	reader, _ := NewReader(db, registry, ReaderOptions{Schema: schema})
	value, err := reader.Capability(ctx, learnerID, "M1-01", now)
	if err != nil {
		t.Fatal(err)
	}
	if value.Capability.Version != 2 || value.Snapshot != nil || len(value.RecentEvidence) != 0 {
		t.Fatalf("current capability read = %#v", value)
	}
	next, err := reader.Next(ctx, learnerID, now)
	if err != nil {
		t.Fatal(err)
	}
	if next == nil || next.Kind != "acquisition" || next.Activity.ID != "guided-run-model" || next.TargetCapability == nil || next.TargetCapability.Version != 2 {
		t.Fatalf("next after version change = %#v", next)
	}
}
