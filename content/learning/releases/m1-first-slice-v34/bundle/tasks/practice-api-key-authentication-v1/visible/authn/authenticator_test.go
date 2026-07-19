package authn

import "testing"

func TestAuthenticatorUsesStrictBearerAndDigestComparison(t *testing.T) {
	authenticator, err := New([]Credential{{Subject: "alice", Token: "alice-secret"}, {Subject: "bob", Token: "bob-secret"}})
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name, header, subject string
		ok                    bool
	}{
		{name: "alice", header: "Bearer alice-secret", subject: "alice", ok: true},
		{name: "bob", header: "Bearer bob-secret", subject: "bob", ok: true},
		{name: "wrong scheme", header: "Basic alice-secret"},
		{name: "extra field", header: "Bearer alice-secret extra"},
		{name: "unknown", header: "Bearer unknown"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			subject, ok := authenticator.Authenticate(test.header)
			if subject != test.subject || ok != test.ok {
				t.Fatalf("Authenticate() = %q, %v", subject, ok)
			}
		})
	}
}

func TestAuthenticatorRejectsUnsafeConfig(t *testing.T) {
	for _, credentials := range [][]Credential{nil, {{Subject: "", Token: "secret"}}, {{Subject: "alice", Token: ""}}, {{Subject: "alice", Token: "same"}, {Subject: "bob", Token: "same"}}} {
		if _, err := New(credentials); err == nil {
			t.Fatalf("New(%#v) succeeded", credentials)
		}
	}
}
