package session

import (
	"errors"
	"time"
)

var (
	ErrNotFound        = errors.New("learning session not found")
	ErrUnauthenticated = errors.New("learning session is not authenticated")
	ErrTokenCollision  = errors.New("learning session token collision")
)

type Session struct {
	ID         string
	LearnerID  string
	CreatedAt  time.Time
	ExpiresAt  time.Time
	LastUsedAt time.Time
}

type Establishment struct {
	Session Session
	Token   string
	Created bool
}

type NewSession struct {
	ID        string
	LearnerID string
	TokenHash string
	CreatedAt time.Time
	ExpiresAt time.Time
}
