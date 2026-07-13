package session

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestServiceEstablishesAndReusesHashedSession(t *testing.T) {
	now := time.Date(2026, time.July, 13, 8, 0, 0, 0, time.UTC)
	repository := newMemoryRepository()
	service := newTestService(t, repository, now)

	created, err := service.Establish(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}
	if !created.Created || len(created.Token) != 43 {
		t.Fatalf("Establish() = %#v, want new 256-bit base64url token", created)
	}
	if got := repository.created[0].TokenHash; got != TokenHash(created.Token) || got == created.Token {
		t.Fatalf("stored token = %q, want SHA-256 hash only", got)
	}
	if created.Session.ExpiresAt != now.Add(24*time.Hour) {
		t.Fatalf("ExpiresAt = %s, want %s", created.Session.ExpiresAt, now.Add(24*time.Hour))
	}

	reused, err := service.Establish(context.Background(), created.Token)
	if err != nil {
		t.Fatal(err)
	}
	if reused.Created || reused.Token != "" || reused.Session.LearnerID != created.Session.LearnerID {
		t.Fatalf("reused Establish() = %#v", reused)
	}
	if len(repository.created) != 1 {
		t.Fatalf("created sessions = %d, want 1", len(repository.created))
	}
}

func TestServiceSeparatesEstablishmentFromAuthentication(t *testing.T) {
	now := time.Date(2026, time.July, 13, 8, 0, 0, 0, time.UTC)
	repository := newMemoryRepository()
	service := newTestService(t, repository, now)

	if _, err := service.Authenticate(context.Background(), ""); !errors.Is(err, ErrUnauthenticated) {
		t.Fatalf("Authenticate(missing) error = %v", err)
	}
	if _, err := service.Authenticate(context.Background(), "forged-token"); !errors.Is(err, ErrUnauthenticated) {
		t.Fatalf("Authenticate(forged) error = %v", err)
	}
	created, err := service.Establish(context.Background(), "forged-token")
	if err != nil {
		t.Fatal(err)
	}
	if !created.Created {
		t.Fatal("Establish(forged) did not create a replacement session")
	}

	repository.sessions[TokenHash(created.Token)] = Session{
		ID: created.Session.ID, LearnerID: created.Session.LearnerID,
		CreatedAt: now.Add(-2 * time.Hour), ExpiresAt: now.Add(-time.Hour), LastUsedAt: now.Add(-time.Hour),
	}
	replaced, err := service.Establish(context.Background(), created.Token)
	if err != nil {
		t.Fatal(err)
	}
	if !replaced.Created || replaced.Session.LearnerID == created.Session.LearnerID {
		t.Fatalf("Establish(expired) = %#v, want new learner session", replaced)
	}
	if _, err := service.Authenticate(context.Background(), created.Token); !errors.Is(err, ErrUnauthenticated) {
		t.Fatalf("Authenticate(expired) error = %v", err)
	}
}

func TestServiceRetriesTokenCollisionWithoutExposingToken(t *testing.T) {
	now := time.Date(2026, time.July, 13, 8, 0, 0, 0, time.UTC)
	repository := newMemoryRepository()
	repository.createErrors = []error{ErrTokenCollision}
	service := newTestService(t, repository, now)

	result, err := service.Establish(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}
	if !result.Created || repository.createCalls != 2 {
		t.Fatalf("Establish() created=%v calls=%d, want retry", result.Created, repository.createCalls)
	}

	repository.findError = errors.New("database unavailable")
	secret := "never-log-this-token"
	_, err = service.Authenticate(context.Background(), secret)
	if err == nil || strings.Contains(err.Error(), secret) {
		t.Fatalf("Authenticate() error = %v, token must not be exposed", err)
	}
}

func newTestService(t *testing.T, repository Repository, now time.Time) *Service {
	t.Helper()
	randomBytes := make([]byte, 1024)
	for index := range randomBytes {
		randomBytes[index] = byte(index)
	}
	service, err := NewService(repository, ServiceOptions{
		TTL: 24 * time.Hour, Random: bytes.NewReader(randomBytes), Now: func() time.Time { return now },
	})
	if err != nil {
		t.Fatal(err)
	}
	return service
}

type memoryRepository struct {
	sessions     map[string]Session
	created      []NewSession
	createErrors []error
	createCalls  int
	findError    error
}

func newMemoryRepository() *memoryRepository {
	return &memoryRepository{sessions: make(map[string]Session)}
}

func (r *memoryRepository) FindActive(_ context.Context, tokenHash string, now time.Time) (Session, error) {
	if r.findError != nil {
		return Session{}, r.findError
	}
	active, exists := r.sessions[tokenHash]
	if !exists || !active.ExpiresAt.After(now) {
		return Session{}, ErrNotFound
	}
	active.LastUsedAt = now
	r.sessions[tokenHash] = active
	return active, nil
}

func (r *memoryRepository) Create(_ context.Context, input NewSession) (Session, error) {
	r.createCalls++
	if len(r.createErrors) > 0 {
		err := r.createErrors[0]
		r.createErrors = r.createErrors[1:]
		return Session{}, err
	}
	r.created = append(r.created, input)
	created := Session{
		ID: input.ID, LearnerID: input.LearnerID, CreatedAt: input.CreatedAt,
		ExpiresAt: input.ExpiresAt, LastUsedAt: input.CreatedAt,
	}
	r.sessions[input.TokenHash] = created
	return created, nil
}
