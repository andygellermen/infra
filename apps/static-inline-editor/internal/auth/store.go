package auth

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	ErrTokenNotFound = errors.New("token not found")
	ErrTokenExpired  = errors.New("token expired")
	ErrTokenUsed     = errors.New("token already used")
)

type Store struct {
	path string
	mu   sync.Mutex
}

type MagicLink struct {
	Token     string
	Tenant    string
	Email     string
	CreatedAt time.Time
	ExpiresAt time.Time
}

type Session struct {
	Token     string
	Tenant    string
	Email     string
	CreatedAt time.Time
	ExpiresAt time.Time
}

type record struct {
	Token      string     `json:"token"`
	Kind       string     `json:"kind"`
	Tenant     string     `json:"tenant"`
	Email      string     `json:"email"`
	CreatedAt  time.Time  `json:"created_at"`
	ExpiresAt  time.Time  `json:"expires_at"`
	ConsumedAt *time.Time `json:"consumed_at,omitempty"`
}

type state struct {
	Records []record `json:"records"`
}

func NewStore(path string) *Store {
	return &Store{path: path}
}

func (s *Store) CreateMagicLink(tenant, email, ttl string) (string, error) {
	duration, err := time.ParseDuration(ttl)
	if err != nil {
		return "", fmt.Errorf("parse magic-link ttl: %w", err)
	}
	token, err := randomToken(32)
	if err != nil {
		return "", err
	}

	now := time.Now().UTC()
	item := record{
		Token:     token,
		Kind:      "magic_link",
		Tenant:    tenant,
		Email:     email,
		CreatedAt: now,
		ExpiresAt: now.Add(duration),
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := s.loadLocked()
	if err != nil {
		return "", err
	}
	data.Records = append(cleanRecords(data.Records, now), item)
	if err := s.saveLocked(data); err != nil {
		return "", err
	}
	return token, nil
}

func (s *Store) ConsumeMagicLink(token string) (MagicLink, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	data, err := s.loadLocked()
	if err != nil {
		return MagicLink{}, err
	}

	for idx, item := range data.Records {
		if item.Kind != "magic_link" || item.Token != token {
			continue
		}
		if item.ConsumedAt != nil {
			return MagicLink{}, ErrTokenUsed
		}
		if now.After(item.ExpiresAt) {
			return MagicLink{}, ErrTokenExpired
		}
		data.Records[idx].ConsumedAt = &now
		data.Records = cleanRecords(data.Records, now)
		if err := s.saveLocked(data); err != nil {
			return MagicLink{}, err
		}
		return MagicLink{
			Token:     item.Token,
			Tenant:    item.Tenant,
			Email:     item.Email,
			CreatedAt: item.CreatedAt,
			ExpiresAt: item.ExpiresAt,
		}, nil
	}

	return MagicLink{}, ErrTokenNotFound
}

func (s *Store) CreateSession(tenant, email, ttl string) (string, error) {
	duration, err := time.ParseDuration(ttl)
	if err != nil {
		return "", fmt.Errorf("parse session ttl: %w", err)
	}
	token, err := randomToken(32)
	if err != nil {
		return "", err
	}

	now := time.Now().UTC()
	item := record{
		Token:     token,
		Kind:      "session",
		Tenant:    tenant,
		Email:     email,
		CreatedAt: now,
		ExpiresAt: now.Add(duration),
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := s.loadLocked()
	if err != nil {
		return "", err
	}
	data.Records = append(cleanRecords(data.Records, now), item)
	if err := s.saveLocked(data); err != nil {
		return "", err
	}
	return token, nil
}

func (s *Store) GetSession(token string) (Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	data, err := s.loadLocked()
	if err != nil {
		return Session{}, err
	}

	for _, item := range data.Records {
		if item.Kind != "session" || item.Token != token {
			continue
		}
		if item.ConsumedAt != nil {
			return Session{}, ErrTokenUsed
		}
		if now.After(item.ExpiresAt) {
			return Session{}, ErrTokenExpired
		}
		return Session{
			Token:     item.Token,
			Tenant:    item.Tenant,
			Email:     item.Email,
			CreatedAt: item.CreatedAt,
			ExpiresAt: item.ExpiresAt,
		}, nil
	}

	return Session{}, ErrTokenNotFound
}

func (s *Store) DeleteSession(token string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	data, err := s.loadLocked()
	if err != nil {
		return err
	}

	filtered := make([]record, 0, len(data.Records))
	for _, item := range cleanRecords(data.Records, now) {
		if item.Kind == "session" && item.Token == token {
			continue
		}
		filtered = append(filtered, item)
	}
	data.Records = filtered
	return s.saveLocked(data)
}

func (s *Store) loadLocked() (state, error) {
	raw, err := os.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return state{}, nil
		}
		return state{}, fmt.Errorf("read auth store: %w", err)
	}

	var data state
	if len(raw) == 0 {
		return state{}, nil
	}
	if err := json.Unmarshal(raw, &data); err != nil {
		return state{}, fmt.Errorf("decode auth store: %w", err)
	}
	return data, nil
}

func (s *Store) saveLocked(data state) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return fmt.Errorf("create auth store dir: %w", err)
	}

	raw, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("encode auth store: %w", err)
	}

	tmpPath := s.path + ".tmp"
	if err := os.WriteFile(tmpPath, raw, 0o600); err != nil {
		return fmt.Errorf("write auth store temp file: %w", err)
	}
	if err := os.Rename(tmpPath, s.path); err != nil {
		return fmt.Errorf("rename auth store temp file: %w", err)
	}
	return nil
}

func cleanRecords(records []record, now time.Time) []record {
	out := make([]record, 0, len(records))
	for _, item := range records {
		if now.After(item.ExpiresAt) {
			continue
		}
		if item.Kind == "magic_link" && item.ConsumedAt != nil {
			continue
		}
		out = append(out, item)
	}
	return out
}

func randomToken(size int) (string, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	return hex.EncodeToString(buf), nil
}
