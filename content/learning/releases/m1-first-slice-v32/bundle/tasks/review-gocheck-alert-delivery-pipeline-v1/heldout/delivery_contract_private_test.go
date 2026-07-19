package deliverycontract_test

import (
	_ "embed"
	"strings"
	"testing"

	"gocheckalerts/internal/alertdelivery"
)

//go:embed Dockerfile
var dockerfile string

//go:embed Makefile
var makefile string

//go:embed .github/workflows/ci.yml
var workflow string

func requireTerms(t *testing.T, source string, terms ...string) {
	t.Helper()
	for _, term := range terms {
		if !strings.Contains(source, term) {
			t.Fatalf("missing %q", term)
		}
	}
}

func TestMigrationPolicyRejectsDownAndChecksumDrift(t *testing.T) {
	valid := []alertdelivery.Migration{{Name: "0001_create_alerts.up.sql", SHA256: strings.Repeat("a", 64)}, {Name: "0002_add_attempts.up.sql", SHA256: strings.Repeat("b", 64)}}
	if err := alertdelivery.ValidateManifest(valid); err != nil {
		t.Fatal(err)
	}
	for _, entries := range [][]alertdelivery.Migration{{{Name: "0001_alerts.down.sql", SHA256: strings.Repeat("a", 64)}}, {{Name: "0001_alerts.up.sql", SHA256: "bad"}}, {{Name: "0002_b.up.sql", SHA256: strings.Repeat("b", 64)}, {Name: "0001_a.up.sql", SHA256: strings.Repeat("a", 64)}}, {{Name: "0001_a.up.sql", SHA256: strings.Repeat("a", 64)}, {Name: "0001_b.up.sql", SHA256: strings.Repeat("b", 64)}}} {
		if err := alertdelivery.ValidateManifest(entries); err == nil {
			t.Fatalf("accepted %#v", entries)
		}
	}
}

func TestDockerfileUsesMinimalNonRootRuntime(t *testing.T) {
	requireTerms(t, dockerfile, "FROM golang:1.25-alpine AS builder", "FROM alpine:3.22", "CGO_ENABLED=0", "-trimpath", "./cmd/alertworker", "COPY --from=builder", "adduser", "USER worker", "ENTRYPOINT")
	if strings.Count(dockerfile, "FROM ") != 2 {
		t.Fatalf("want two stages: %s", dockerfile)
	}
}

func TestCICoversRequiredDeliveryGates(t *testing.T) {
	requireTerms(t, workflow, "actions/checkout@v4", "actions/setup-go@v5", "go-version: '1.25'", "gofmt -l", "go vet ./...", "go test ./...", "go test -race ./...", "govulncheck ./...", "migration-check", "docker build")
}

func TestMakefileMirrorsCI(t *testing.T) {
	requireTerms(t, makefile, "fmt-check:", "vet:", "test:", "race:", "vuln:", "migration-check:", "image:", "verify:")
}
