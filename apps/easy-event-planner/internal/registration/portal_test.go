package registration

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/andygellermann/infra/apps/easy-event-planner/internal/event"
	"github.com/andygellermann/infra/apps/easy-event-planner/internal/tenant"
)

func createConfirmedRegistrationForPortalTest(t *testing.T, service *Service, tenantID, tenantSlug string, eventItem event.Event, name, email string) StartResult {
	t.Helper()

	start, err := service.Start(context.Background(), StartInput{
		TenantID:          tenantID,
		TenantSlug:        tenantSlug,
		EventID:           eventItem.ID,
		Name:              name,
		Email:             email,
		ParticipationType: event.ParticipationModeOnsite,
		PrivacyAccepted:   true,
	})
	if err != nil {
		t.Fatalf("start registration: %v", err)
	}
	token := extractVerifyTokenFromLatestEmailJob(t, service.db, tenantID, start.RegistrationID)
	if _, err := service.Verify(context.Background(), VerifyInput{TenantID: tenantID, RawToken: token}); err != nil {
		t.Fatalf("verify registration: %v", err)
	}
	return start
}

func TestListParticipantRegistrations(t *testing.T) {
	service, dbHandle, tenantItem := setupRegistrationService(t)
	eventItem := createPublishedEventForRegistration(t, dbHandle, tenantItem.ID, event.CreateEventParams{
		Slug:     "portal-list-event",
		Title:    "Portal List Event",
		StartsAt: time.Now().UTC().Add(48 * time.Hour).Format(time.RFC3339),
	})
	start := createConfirmedRegistrationForPortalTest(t, service, tenantItem.ID, tenantItem.Slug, eventItem, "Portal User", "portal-user@example.com")

	items, err := service.ListParticipantRegistrations(context.Background(), tenantItem.ID, start.ParticipantID)
	if err != nil {
		t.Fatalf("list participant registrations: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected exactly one participant registration, got %d", len(items))
	}
	if items[0].ID != start.RegistrationID {
		t.Fatalf("expected registration id %q, got %q", start.RegistrationID, items[0].ID)
	}
	if items[0].Status != StatusConfirmed {
		t.Fatalf("expected status confirmed, got %q", items[0].Status)
	}
	if items[0].EventSlug != eventItem.Slug {
		t.Fatalf("expected event slug %q, got %q", eventItem.Slug, items[0].EventSlug)
	}
}

func TestCancelParticipantRegistration(t *testing.T) {
	service, dbHandle, tenantItem := setupRegistrationService(t)
	eventItem := createPublishedEventForRegistration(t, dbHandle, tenantItem.ID, event.CreateEventParams{
		Slug:     "portal-cancel-event",
		Title:    "Portal Cancel Event",
		StartsAt: time.Now().UTC().Add(72 * time.Hour).Format(time.RFC3339),
	})
	start := createConfirmedRegistrationForPortalTest(t, service, tenantItem.ID, tenantItem.Slug, eventItem, "Portal Cancel", "portal-cancel@example.com")

	cancelled, err := service.CancelParticipantRegistration(context.Background(), tenantItem.ID, start.ParticipantID, start.RegistrationID, "Ich kann leider nicht teilnehmen")
	if err != nil {
		t.Fatalf("cancel participant registration: %v", err)
	}
	if cancelled.Status != StatusCancelled {
		t.Fatalf("expected cancelled status, got %q", cancelled.Status)
	}
	if cancelled.CancelledAt == nil {
		t.Fatalf("expected cancelled_at to be set")
	}

	_, err = service.CancelParticipantRegistration(context.Background(), tenantItem.ID, start.ParticipantID, start.RegistrationID, "zweiter Versuch")
	if !errors.Is(err, ErrRegistrationCancelNotAllowed) {
		t.Fatalf("expected ErrRegistrationCancelNotAllowed on second cancel, got %v", err)
	}
}

func TestCancelParticipantRegistrationDeadlineExceeded(t *testing.T) {
	service, dbHandle, tenantItem := setupRegistrationService(t)
	tenantRepo := tenant.NewRepository(dbHandle)
	if _, err := tenantRepo.UpsertSettings(context.Background(), tenant.UpsertTenantSettingsParams{
		TenantID: tenantItem.ID,
		Settings: tenant.TenantSettingsInput{
			SettingsJSON: `{"participant_cancel_deadline_hours":24}`,
		},
	}); err != nil {
		t.Fatalf("upsert tenant settings: %v", err)
	}

	eventItem := createPublishedEventForRegistration(t, dbHandle, tenantItem.ID, event.CreateEventParams{
		Slug:     "portal-cancel-too-late",
		Title:    "Portal Cancel Too Late",
		StartsAt: time.Now().UTC().Add(12 * time.Hour).Format(time.RFC3339),
	})
	start := createConfirmedRegistrationForPortalTest(t, service, tenantItem.ID, tenantItem.Slug, eventItem, "Portal Late", "portal-late@example.com")

	item, err := service.GetParticipantRegistration(context.Background(), tenantItem.ID, start.ParticipantID, start.RegistrationID)
	if err != nil {
		t.Fatalf("get participant registration: %v", err)
	}
	if item.SelfCancelAllowed {
		t.Fatalf("expected self cancel to be blocked by deadline")
	}
	if item.SelfCancelDeadline == nil {
		t.Fatalf("expected self cancel deadline to be set")
	}
	if item.SelfCancelDeadlineHours != 24 {
		t.Fatalf("expected self cancel deadline hours 24, got %d", item.SelfCancelDeadlineHours)
	}

	_, err = service.CancelParticipantRegistration(context.Background(), tenantItem.ID, start.ParticipantID, start.RegistrationID, "Zu spaet")
	if !errors.Is(err, ErrRegistrationCancelDeadlineExceeded) {
		t.Fatalf("expected ErrRegistrationCancelDeadlineExceeded, got %v", err)
	}
}

func TestParticipantRegistrationAccessDenied(t *testing.T) {
	service, dbHandle, tenantItem := setupRegistrationService(t)
	eventItem := createPublishedEventForRegistration(t, dbHandle, tenantItem.ID, event.CreateEventParams{
		Slug:     "portal-access-event",
		Title:    "Portal Access Event",
		StartsAt: time.Now().UTC().Add(96 * time.Hour).Format(time.RFC3339),
	})
	ownerStart := createConfirmedRegistrationForPortalTest(t, service, tenantItem.ID, tenantItem.Slug, eventItem, "Owner", "owner-participant@example.com")
	otherStart := createConfirmedRegistrationForPortalTest(t, service, tenantItem.ID, tenantItem.Slug, eventItem, "Other", "other-participant@example.com")

	_, err := service.GetParticipantRegistration(context.Background(), tenantItem.ID, otherStart.ParticipantID, ownerStart.RegistrationID)
	if !errors.Is(err, ErrParticipantAccessDenied) {
		t.Fatalf("expected ErrParticipantAccessDenied, got %v", err)
	}
}
