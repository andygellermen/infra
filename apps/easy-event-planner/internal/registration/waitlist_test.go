package registration

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/andygellermann/infra/apps/easy-event-planner/internal/event"
)

func TestWaitlistListOfferPromoteFlow(t *testing.T) {
	service, dbHandle, tenantItem := setupRegistrationService(t)
	maxOne := 1
	waitlistEnabled := true
	eventItem := createPublishedEventForRegistration(t, dbHandle, tenantItem.ID, event.CreateEventParams{
		Slug:            "waitlist-admin-flow",
		Title:           "Waitlist Admin Flow",
		StartsAt:        time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339),
		MaxParticipants: &maxOne,
		WaitlistEnabled: &waitlistEnabled,
	})

	firstStart, err := service.Start(context.Background(), StartInput{
		TenantID:        tenantItem.ID,
		TenantSlug:      tenantItem.Slug,
		EventID:         eventItem.ID,
		Name:            "Alice",
		Email:           "alice@example.com",
		PrivacyAccepted: true,
	})
	if err != nil {
		t.Fatalf("first Start returned error: %v", err)
	}
	firstToken := extractVerifyTokenFromLatestEmailJob(t, dbHandle, tenantItem.ID, firstStart.RegistrationID)
	if _, err := service.Verify(context.Background(), VerifyInput{TenantID: tenantItem.ID, RawToken: firstToken}); err != nil {
		t.Fatalf("first Verify returned error: %v", err)
	}

	secondStart, err := service.Start(context.Background(), StartInput{
		TenantID:        tenantItem.ID,
		TenantSlug:      tenantItem.Slug,
		EventID:         eventItem.ID,
		Name:            "Bob",
		Email:           "bob@example.com",
		PrivacyAccepted: true,
	})
	if err != nil {
		t.Fatalf("second Start returned error: %v", err)
	}
	secondToken := extractVerifyTokenFromLatestEmailJob(t, dbHandle, tenantItem.ID, secondStart.RegistrationID)
	waitlistResult, err := service.Verify(context.Background(), VerifyInput{TenantID: tenantItem.ID, RawToken: secondToken})
	if err != nil {
		t.Fatalf("second Verify returned error: %v", err)
	}
	if waitlistResult.Status != StatusWaitlist {
		t.Fatalf("expected waitlist status, got %q", waitlistResult.Status)
	}
	waitlistID := waitlistResult.WaitlistID
	if waitlistID == "" {
		t.Fatalf("expected waitlist id")
	}

	listed, err := service.ListWaitlistEntries(context.Background(), tenantItem.ID, eventItem.ID)
	if err != nil {
		t.Fatalf("ListWaitlistEntries returned error: %v", err)
	}
	if len(listed) != 1 {
		t.Fatalf("expected one waitlist entry, got %d", len(listed))
	}
	if listed[0].Status != WaitlistStatusWaiting {
		t.Fatalf("expected waiting status, got %q", listed[0].Status)
	}

	offered, err := service.OfferWaitlistEntry(context.Background(), tenantItem.ID, waitlistID, "127.0.0.1", "unit-test")
	if err != nil {
		t.Fatalf("OfferWaitlistEntry returned error: %v", err)
	}
	if offered.Status != WaitlistStatusOffered {
		t.Fatalf("expected offered status, got %q", offered.Status)
	}
	if offered.OfferedAt == nil || offered.OfferExpiresAt == nil {
		t.Fatalf("expected offered_at and offer_expires_at to be set")
	}

	_, err = service.PromoteWaitlistEntry(context.Background(), tenantItem.ID, waitlistID)
	if !errors.Is(err, ErrEventFull) {
		t.Fatalf("expected ErrEventFull while event is full, got %v", err)
	}

	if _, err := dbHandle.ExecContext(
		context.Background(),
		`UPDATE registrations
     SET status = ?, cancelled_at = ?, updated_at = ?
     WHERE id = ?`,
		StatusCancelled,
		time.Now().UTC().Format(time.RFC3339),
		time.Now().UTC().Format(time.RFC3339),
		firstStart.RegistrationID,
	); err != nil {
		t.Fatalf("cancel first registration: %v", err)
	}

	promoted, err := service.PromoteWaitlistEntry(context.Background(), tenantItem.ID, waitlistID)
	if err != nil {
		t.Fatalf("PromoteWaitlistEntry returned error: %v", err)
	}
	if promoted.Status != WaitlistStatusPromoted {
		t.Fatalf("expected promoted status, got %q", promoted.Status)
	}
	if promoted.AcceptedAt == nil {
		t.Fatalf("expected accepted_at to be set")
	}
	if promoted.RegistrationStatus != StatusConfirmed {
		t.Fatalf("expected registration status confirmed, got %q", promoted.RegistrationStatus)
	}

	_, err = service.PromoteWaitlistEntry(context.Background(), tenantItem.ID, waitlistID)
	if !errors.Is(err, ErrWaitlistStateInvalid) {
		t.Fatalf("expected ErrWaitlistStateInvalid on second promote, got %v", err)
	}
}

func TestOfferWaitlistEntryNotFound(t *testing.T) {
	service, _, tenantItem := setupRegistrationService(t)
	_, err := service.OfferWaitlistEntry(context.Background(), tenantItem.ID, "missing-entry", "127.0.0.1", "unit-test")
	if !errors.Is(err, ErrWaitlistEntryNotFound) {
		t.Fatalf("expected ErrWaitlistEntryNotFound, got %v", err)
	}
}
