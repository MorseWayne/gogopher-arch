package session

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"time"
)

const (
	tokenBytes        = 32
	createMaxAttempts = 3
)

type Repository interface {
	FindActive(context.Context, string, time.Time) (Session, error)
	Create(context.Context, NewSession) (Session, error)
}

type ServiceOptions struct {
	TTL    time.Duration
	Random io.Reader
	Now    func() time.Time
}

type Service struct {
	repository Repository
	ttl        time.Duration
	random     io.Reader
	now        func() time.Time
}

func NewService(repository Repository, options ServiceOptions) (*Service, error) {
	if repository == nil {
		return nil, fmt.Errorf("session repository is required")
	}
	if options.TTL <= 0 {
		return nil, fmt.Errorf("session TTL must be positive")
	}
	if options.Random == nil {
		options.Random = rand.Reader
	}
	if options.Now == nil {
		options.Now = time.Now
	}
	return &Service{repository: repository, ttl: options.TTL, random: options.Random, now: options.Now}, nil
}

func (s *Service) Establish(ctx context.Context, token string) (Establishment, error) {
	now := s.now().UTC()
	if token != "" {
		active, err := s.repository.FindActive(ctx, TokenHash(token), now)
		if err == nil {
			return Establishment{Session: active}, nil
		}
		if !errors.Is(err, ErrNotFound) {
			return Establishment{}, fmt.Errorf("find active learning session: %w", err)
		}
	}

	for attempt := 0; attempt < createMaxAttempts; attempt++ {
		rawToken, err := randomToken(s.random)
		if err != nil {
			return Establishment{}, err
		}
		learnerID, err := randomUUID(s.random)
		if err != nil {
			return Establishment{}, err
		}
		sessionID, err := randomUUID(s.random)
		if err != nil {
			return Establishment{}, err
		}
		created, err := s.repository.Create(ctx, NewSession{
			ID: sessionID, LearnerID: learnerID, TokenHash: TokenHash(rawToken),
			CreatedAt: now, ExpiresAt: now.Add(s.ttl),
		})
		if err == nil {
			return Establishment{Session: created, Token: rawToken, Created: true}, nil
		}
		if !errors.Is(err, ErrTokenCollision) {
			return Establishment{}, fmt.Errorf("create learning session: %w", err)
		}
	}
	return Establishment{}, fmt.Errorf("create learning session after %d token collisions", createMaxAttempts)
}

func (s *Service) Authenticate(ctx context.Context, token string) (Session, error) {
	if token == "" {
		return Session{}, ErrUnauthenticated
	}
	active, err := s.repository.FindActive(ctx, TokenHash(token), s.now().UTC())
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return Session{}, ErrUnauthenticated
		}
		return Session{}, fmt.Errorf("authenticate learning session: %w", err)
	}
	return active, nil
}

func TokenHash(token string) string {
	digest := sha256.Sum256([]byte(token))
	return hex.EncodeToString(digest[:])
}

func randomToken(source io.Reader) (string, error) {
	buffer := make([]byte, tokenBytes)
	if _, err := io.ReadFull(source, buffer); err != nil {
		return "", fmt.Errorf("generate learning session token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buffer), nil
}

func randomUUID(source io.Reader) (string, error) {
	var value [16]byte
	if _, err := io.ReadFull(source, value[:]); err != nil {
		return "", fmt.Errorf("generate UUID: %w", err)
	}
	value[6] = (value[6] & 0x0f) | 0x40
	value[8] = (value[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		value[0:4], value[4:6], value[6:8], value[8:10], value[10:16]), nil
}
