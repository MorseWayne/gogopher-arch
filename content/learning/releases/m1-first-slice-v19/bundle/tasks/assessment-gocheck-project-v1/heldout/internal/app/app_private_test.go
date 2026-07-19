package app

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gocheck/internal/check"
)

func TestRunUsesFlagsAndExitCodes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if strings.HasSuffix(request.URL.Path, "/down") {
			writer.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		writer.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	writeConfig := func(body string) string {
		path := filepath.Join(t.TempDir(), "targets.json")
		if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
		return path
	}
	dependencies := Dependencies{Client: check.Client(server.Client())}

	success := writeConfig(fmt.Sprintf(`{"targets":[{"name":"api","url":%q}]}`, server.URL+"/up"))
	var stdout, stderr bytes.Buffer
	if code := Run(context.Background(), []string{"-config", success, "-format", "text", "-concurrency", "1"}, &stdout, &stderr, dependencies); code != 0 || stdout.String() != "api\tok\t204\n" {
		t.Fatalf("Run(success) code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}

	partial := writeConfig(fmt.Sprintf(`{"targets":[{"name":"api","url":%q},{"name":"db","url":%q}]}`, server.URL+"/up", server.URL+"/down"))
	stdout.Reset()
	stderr.Reset()
	if code := Run(context.Background(), []string{"-config", partial, "-format", "json"}, &stdout, &stderr, dependencies); code != 1 || !strings.Contains(stdout.String(), `"status_code":503`) {
		t.Fatalf("Run(partial) code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	if code := Run(context.Background(), []string{"-config", filepath.Join(t.TempDir(), "missing.json")}, &stdout, &stderr, dependencies); code != 2 || stderr.Len() == 0 {
		t.Fatalf("Run(config error) code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
}
