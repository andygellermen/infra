package registration

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/andygellermann/infra/apps/easy-event-planner/internal/event"
)

func startAndVerifyRegistrationForAdminTest(
	t *testing.T,
	service *Service,
	tenantID string,
	input StartInput,
) (StartResult, VerifyResult) {
	t.Helper()

	start, err := service.Start(context.Background(), input)
	if err != nil {
		t.Fatalf("Start returned error: %v", err)
	}
	token := extractVerifyTokenFromLatestEmailJob(t, service.db, tenantID, start.RegistrationID)
	verify, err := service.Verify(context.Background(), VerifyInput{
		TenantID: tenantID,
		RawToken: token,
	})
	if err != nil {
		t.Fatalf("Verify returned error: %v", err)
	}
	return start, verify
}

func TestListEventRegistrationsAndGetRegistration(t *testing.T) {
	service, dbHandle, tenantItem := setupRegistrationService(t)

	eventItem := createPublishedEventForRegistration(t, dbHandle, tenantItem.ID, event.CreateEventParams{
		Slug:     "admin-registrations",
		Title:    "Admin Registrations",
		StartsAt: time.Now().UTC().Add(48 * time.Hour).Format(time.RFC3339),
	})

	aliceStart, aliceVerify := startAndVerifyRegistrationForAdminTest(t, service, tenantItem.ID, StartInput{
		TenantID:        tenantItem.ID,
		TenantSlug:      tenantItem.Slug,
		EventID:         eventItem.ID,
		Name:            "Alice",
		Email:           "alice@example.com",
		PrivacyAccepted: true,
	})
	if aliceVerify.Status != StatusConfirmed {
		t.Fatalf("expected confirmed status, got %q", aliceVerify.Status)
	}

	bobStart, err := service.Start(context.Background(), StartInput{
		TenantID:        tenantItem.ID,
		TenantSlug:      tenantItem.Slug,
		EventID:         eventItem.ID,
		Name:            "Bob",
		Email:           "bob@example.com",
		PrivacyAccepted: true,
	})
	if err != nil {
		t.Fatalf("Start for Bob returned error: %v", err)
	}

	items, err := service.ListEventRegistrations(context.Background(), tenantItem.ID, eventItem.ID)
	if err != nil {
		t.Fatalf("ListEventRegistrations returned error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 registrations, got %d", len(items))
	}
	byID := make(map[string]AdminRegistration, len(items))
	for _, item := range items {
		byID[item.ID] = item
	}

	bobItem, ok := byID[bobStart.RegistrationID]
	if !ok {
		t.Fatalf("expected bob registration %s in list", bobStart.RegistrationID)
	}
	if bobItem.Status != StatusVerificationPending {
		t.Fatalf("expected bob status verification_pending, got %q", bobItem.Status)
	}

	aliceItem, ok := byID[aliceStart.RegistrationID]
	if !ok {
		t.Fatalf("expected alice registration %s in list", aliceStart.RegistrationID)
	}
	if aliceItem.Status != StatusConfirmed {
		t.Fatalf("expected alice status confirmed, got %q", aliceItem.Status)
	}
	if aliceItem.ConfirmedAt == nil {
		t.Fatalf("expected confirmed_at for alice registration")
	}

	detail, err := service.GetRegistration(context.Background(), tenantItem.ID, aliceStart.RegistrationID)
	if err != nil {
		t.Fatalf("GetRegistration returned error: %v", err)
	}
	if detail.ID != aliceStart.RegistrationID {
		t.Fatalf("expected registration id %s, got %s", aliceStart.RegistrationID, detail.ID)
	}
	if detail.ParticipantEmail != "alice@example.com" {
		t.Fatalf("expected participant email alice@example.com, got %q", detail.ParticipantEmail)
	}
}

func TestListEventRegistrationsReturnsEventNotFound(t *testing.T) {
	service, _, tenantItem := setupRegistrationService(t)

	_, err := service.ListEventRegistrations(context.Background(), tenantItem.ID, "evt_missing")
	if !errors.Is(err, ErrEventNotFound) {
		t.Fatalf("expected ErrEventNotFound, got %v", err)
	}
}

func TestGetDashboardSummary(t *testing.T) {
	service, dbHandle, tenantItem := setupRegistrationService(t)
	now := time.Date(2026, 1, 15, 9, 0, 0, 0, time.UTC)
	service.nowFn = func() time.Time { return now }

	maxTwo := 2
	todayEvent := createPublishedEventForRegistration(t, dbHandle, tenantItem.ID, event.CreateEventParams{
		Slug:            "today-event",
		Title:           "Today Event",
		StartsAt:        now.Add(2 * time.Hour).Format(time.RFC3339),
		MaxParticipants: &maxTwo,
	})

	maxOne := 1
	nextEvent := createPublishedEventForRegistration(t, dbHandle, tenantItem.ID, event.CreateEventParams{
		Slug:            "next-event",
		Title:           "Next Event",
		StartsAt:        now.Add(48 * time.Hour).Format(time.RFC3339),
		MaxParticipants: &maxOne,
	})

	_, verifyToday := startAndVerifyRegistrationForAdminTest(t, service, tenantItem.ID, StartInput{
		TenantID:        tenantItem.ID,
		TenantSlug:      tenantItem.Slug,
		EventID:         todayEvent.ID,
		Name:            "Alice",
		Email:           "alice@example.com",
		PrivacyAccepted: true,
	})
	if verifyToday.Status != StatusConfirmed {
		t.Fatalf("expected confirmed for today event, got %q", verifyToday.Status)
	}

	_, verifyNextA := startAndVerifyRegistrationForAdminTest(t, service, tenantItem.ID, StartInput{
		TenantID:        tenantItem.ID,
		TenantSlug:      tenantItem.Slug,
		EventID:         nextEvent.ID,
		Name:            "Bob",
		Email:           "bob@example.com",
		PrivacyAccepted: true,
	})
	if verifyNextA.Status != StatusConfirmed {
		t.Fatalf("expected confirmed for next event, got %q", verifyNextA.Status)
	}

	_, verifyNextB := startAndVerifyRegistrationForAdminTest(t, service, tenantItem.ID, StartInput{
		TenantID:        tenantItem.ID,
		TenantSlug:      tenantItem.Slug,
		EventID:         nextEvent.ID,
		Name:            "Carla",
		Email:           "carla@example.com",
		PrivacyAccepted: true,
	})
	if verifyNextB.Status != StatusWaitlist {
		t.Fatalf("expected waitlist for second next-event registration, got %q", verifyNextB.Status)
	}

	dashboard, err := service.GetDashboard(context.Background(), tenantItem.ID)
	if err != nil {
		t.Fatalf("GetDashboard returned error: %v", err)
	}

	if dashboard.Stats.TodayEvents != 1 {
		t.Fatalf("expected today_events=1, got %d", dashboard.Stats.TodayEvents)
	}
	if dashboard.Stats.UpcomingEvents != 2 {
		t.Fatalf("expected upcoming_events=2, got %d", dashboard.Stats.UpcomingEvents)
	}
	if dashboard.Stats.ConfirmedParticipants != 2 {
		t.Fatalf("expected confirmed_participants=2, got %d", dashboard.Stats.ConfirmedParticipants)
	}
	if dashboard.Stats.WaitlistEntries != 1 {
		t.Fatalf("expected waitlist_entries=1, got %d", dashboard.Stats.WaitlistEntries)
	}
	if dashboard.Stats.FreeSeats != 1 {
		t.Fatalf("expected free_seats=1, got %d", dashboard.Stats.FreeSeats)
	}
	if dashboard.Stats.OpenEmailJobs == 0 {
		t.Fatalf("expected open email jobs > 0")
	}
	if len(dashboard.Today) != 1 {
		t.Fatalf("expected one today event, got %d", len(dashboard.Today))
	}
	if len(dashboard.NextEvents) < 2 {
		t.Fatalf("expected at least two upcoming events, got %d", len(dashboard.NextEvents))
	}
}

func TestGetRegistrationNotFound(t *testing.T) {
	service, _, tenantItem := setupRegistrationService(t)

	_, err := service.GetRegistration(context.Background(), tenantItem.ID, "reg_missing")
	if !errors.Is(err, ErrRegistrationNotFound) {
		t.Fatalf("expected ErrRegistrationNotFound, got %v", err)
	}
}
