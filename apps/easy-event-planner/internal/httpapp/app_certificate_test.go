package httpapp

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/andygellermann/infra/apps/easy-event-planner/internal/event"
)

func TestCertificateAdminParticipantAndPublicFlow(t *testing.T) {
	app, sender, tenantSlug := setupAuthApp(t)
	tenantID := tenantIDBySlug(t, app, tenantSlug)
	adminCookie := loginSessionCookie(t, app, sender, tenantSlug, "owner@example.com")

	eventItem := createPublishedEventForRegistrationHTTP(t, app, tenantID, event.CreateEventParams{
		Slug:     "certificate-http-flow",
		Title:    "Certificate HTTP Flow",
		StartsAt: time.Now().UTC().Add(48 * time.Hour).Format(time.RFC3339),
	})
	registrationID := createConfirmedRegistrationForPaymentHTTPTest(
		t,
		app,
		tenantSlug,
		eventItem.ID,
		"participant-cert@example.com",
	)

	markAttendedReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/registrations/"+registrationID+"/mark-attended", nil)
	markAttendedReq.AddCookie(adminCookie)
	markAttendedRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(markAttendedRec, markAttendedReq)
	if markAttendedRec.Code != http.StatusOK {
		t.Fatalf("expected mark-attended 200, got %d", markAttendedRec.Code)
	}
	markResult := decodeBody[map[string]any](t, markAttendedRec)
	markItem := markResult["item"].(map[string]any)
	if markItem["status"] != "attended" {
		t.Fatalf("expected attended status, got %v", markItem["status"])
	}

	issueReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/registrations/"+registrationID+"/issue-certificate", nil)
	issueReq.AddCookie(adminCookie)
	issueRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(issueRec, issueReq)
	if issueRec.Code != http.StatusCreated {
		t.Fatalf("expected issue-certificate 201, got %d", issueRec.Code)
	}
	issuePayload := decodeBody[map[string]any](t, issueRec)
	certItem := issuePayload["item"].(map[string]any)
	certificateID, _ := certItem["id"].(string)
	if certificateID == "" {
		t.Fatalf("expected certificate id")
	}
	certificateNumber, _ := certItem["certificate_number"].(string)
	if certificateNumber == "" {
		t.Fatalf("expected certificate number")
	}
	verifyURL, _ := certItem["verification_url"].(string)
	if verifyURL == "" {
		t.Fatalf("expected verification_url on first certificate issue")
	}

	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/admin/registrations/"+registrationID+"/certificate", nil)
	getReq.AddCookie(adminCookie)
	getRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("expected get certificate 200, got %d", getRec.Code)
	}
	getPayload := decodeBody[map[string]any](t, getRec)
	getItem := getPayload["item"].(map[string]any)
	if getItem["certificate_number"] != certificateNumber {
		t.Fatalf("expected certificate number %q, got %v", certificateNumber, getItem["certificate_number"])
	}

	portalRequestPayload := map[string]any{
		"email": "participant-cert@example.com",
	}
	portalRequestBody, _ := json.Marshal(portalRequestPayload)
	portalReq := httptest.NewRequest(http.MethodPost, "/api/v1/public/"+tenantSlug+"/participants/portal/request", bytes.NewReader(portalRequestBody))
	portalReq.Header.Set("Content-Type", "application/json")
	portalRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(portalRec, portalReq)
	if portalRec.Code != http.StatusOK {
		t.Fatalf("expected participant portal request 200, got %d", portalRec.Code)
	}

	verifyToken := extractTokenFromVerifyURL(t, sender.lastMessage.VerifyURL)
	portalVerifyBody, _ := json.Marshal(map[string]any{"token": verifyToken})
	portalVerifyReq := httptest.NewRequest(http.MethodPost, "/api/v1/public/"+tenantSlug+"/participants/portal/verify", bytes.NewReader(portalVerifyBody))
	portalVerifyReq.Header.Set("Content-Type", "application/json")
	portalVerifyRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(portalVerifyRec, portalVerifyReq)
	if portalVerifyRec.Code != http.StatusOK {
		t.Fatalf("expected participant portal verify 200, got %d", portalVerifyRec.Code)
	}
	participantCookie := participantSessionCookieFromResponse(t, portalVerifyRec)

	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/public/"+tenantSlug+"/participants/portal/certificates", nil)
	listReq.AddCookie(participantCookie)
	listRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("expected participant certificate list 200, got %d", listRec.Code)
	}
	listPayload := decodeBody[map[string]any](t, listRec)
	items, ok := listPayload["items"].([]any)
	if !ok || len(items) != 1 {
		t.Fatalf("expected one participant certificate item, got %v", listPayload["items"])
	}

	downloadReq := httptest.NewRequest(http.MethodGet, "/api/v1/public/"+tenantSlug+"/participants/portal/certificates/"+certificateID+"/download", nil)
	downloadReq.AddCookie(participantCookie)
	downloadRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(downloadRec, downloadReq)
	if downloadRec.Code != http.StatusOK {
		t.Fatalf("expected participant certificate download 200, got %d", downloadRec.Code)
	}
	if got := downloadRec.Header().Get("Content-Type"); !strings.Contains(got, "application/pdf") {
		t.Fatalf("expected application/pdf content type, got %q", got)
	}
	if !strings.HasPrefix(downloadRec.Body.String(), "%PDF-1.4") {
		t.Fatalf("expected PDF response body header")
	}

	parsedVerifyURL, err := url.Parse(verifyURL)
	if err != nil {
		t.Fatalf("parse verification url: %v", err)
	}
	publicVerifyReq := httptest.NewRequest(http.MethodGet, parsedVerifyURL.RequestURI(), nil)
	publicVerifyRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(publicVerifyRec, publicVerifyReq)
	if publicVerifyRec.Code != http.StatusOK {
		t.Fatalf("expected public verify 200, got %d", publicVerifyRec.Code)
	}
	publicVerifyPayload := decodeBody[map[string]any](t, publicVerifyRec)
	if publicVerifyPayload["valid"] != true {
		t.Fatalf("expected valid=true, got %v", publicVerifyPayload["valid"])
	}
	verifyItem := publicVerifyPayload["item"].(map[string]any)
	if verifyItem["certificate_number"] != certificateNumber {
		t.Fatalf("expected verified certificate number %q, got %v", certificateNumber, verifyItem["certificate_number"])
	}
}
