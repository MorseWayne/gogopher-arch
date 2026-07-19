package deliverycontract_test

import (
	_ "embed"
	"strings"
	"testing"

	"gocheckhub/internal/delivery"
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

func TestMigrationPolicyRejectsDownAndDrift(t *testing.T) {
	valid := []string{"0001_create_checks.up.sql", "0002_add_status.up.sql"}
	if err := delivery.ValidateMigrations(valid); err != nil {
		t.Fatal(err)
	}
	for _, names := range [][]string{{"0001_create_checks.down.sql"}, {"create.up.sql"}, {"0001_a.up.sql", "0001_b.up.sql"}, {"0002_b.up.sql", "0001_a.up.sql"}, {"0001_A.up.sql"}} {
		if err := delivery.ValidateMigrations(names); err == nil {
			t.Fatalf("accepted %#v", names)
		}
	}
}

func TestDockerfileUsesMinimalNonRootRuntime(t *testing.T) {
	requireTerms(t, dockerfile, "FROM golang:1.25-alpine AS builder", "FROM alpine:3.22", "CGO_ENABLED=0", "-trimpath", "COPY --from=builder", "adduser", "USER app", "ENTRYPOINT")
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
