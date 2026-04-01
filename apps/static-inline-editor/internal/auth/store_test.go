package auth

import (
	"errors"
	"path/filepath"
	"testing"
)

func TestStoreMagicLinkAndSession(t *testing.T) {
	store := NewStore(filepath.Join(t.TempDir(), "auth-state.json"))

	magicToken, err := store.CreateMagicLink("example.org", "andy@example.org", "15m")
	if err != nil {
		t.Fatalf("CreateMagicLink returned error: %v", err)
	}

	magicLink, err := store.ConsumeMagicLink(magicToken)
	if err != nil {
		t.Fatalf("ConsumeMagicLink returned error: %v", err)
	}
	if magicLink.Email != "andy@example.org" {
		t.Fatalf("unexpected magic-link email %q", magicLink.Email)
	}

	if _, err := store.ConsumeMagicLink(magicToken); !errors.Is(err, ErrTokenNotFound) && !errors.Is(err, ErrTokenUsed) {
		t.Fatalf("expected magic-link token to be unavailable after consume, got %v", err)
	}

	sessionToken, err := store.CreateSession("example.org", "andy@example.org", "12h")
	if err != nil {
		t.Fatalf("CreateSession returned error: %v", err)
	}

	session, err := store.GetSession(sessionToken)
	if err != nil {
		t.Fatalf("GetSession returned error: %v", err)
	}
	if session.Tenant != "example.org" {
		t.Fatalf("unexpected session tenant %q", session.Tenant)
	}

	if err := store.DeleteSession(sessionToken); err != nil {
		t.Fatalf("DeleteSession returned error: %v", err)
	}
	if _, err := store.GetSession(sessionToken); !errors.Is(err, ErrTokenNotFound) {
		t.Fatalf("expected deleted session to be gone, got %v", err)
	}
}
