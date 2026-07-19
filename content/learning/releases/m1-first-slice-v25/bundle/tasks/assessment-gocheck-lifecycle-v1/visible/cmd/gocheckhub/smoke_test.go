package main

import (
	"context"
	"testing"
)

func TestRunRejectsInvalidConfigurationBeforeInitialization(t *testing.T) {
	called := false
	err := run(context.Background(), Config{}, Dependencies{OpenDatabase: func(context.Context, string) (Database, error) { called = true; return nil, nil }})
	if err == nil || called {
		t.Fatalf("error=%v open called=%v", err, called)
	}
}
