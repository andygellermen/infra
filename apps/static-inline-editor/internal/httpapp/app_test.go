package httpapp

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andygellermann/infra/apps/static-inline-editor/internal/config"
	"github.com/andygellermann/infra/apps/static-inline-editor/internal/model"
)

type fakeMailer struct {
	email string
	url   string
}

func (m *fakeMailer) SendMagicLink(email, verifyURL string) error {
	m.email = email
	m.url = verifyURL
	return nil
}

func TestRequestLinkAndVerifyFlow(t *testing.T) {
	cfg := config.Config{
		Addr:          ":8090",
		DataDir:       t.TempDir(),
		SessionTTL:    "12h",
		MagicLinkTTL:  "15m",
		SecureCookies: false,
		Tenants: map[string]model.Tenant{
			"example.org": {
				Domain:        "example.org",
				LoginDomain:   "bearbeitung.example.org",
				AllowedEmails: []string{"andy@example.org"},
				StartPath:     "/index.html",
			},
		},
	}

	app := New(cfg)
	mailer := &fakeMailer{}
	app.mailer = mailer

	req := httptest.NewRequest(http.MethodPost, "http://bearbeitung.example.org/auth/request-link", strings.NewReader(url.Values{
		"email": []string{"andy@example.org"},
	}.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Host", "bearbeitung.example.org")
	rec := httptest.NewRecorder()
	app.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200 for request-link, got %d", rec.Code)
	}
	if mailer.email != "andy@example.org" {
		t.Fatalf("expected magic-link mail to be addressed to allowed email")
	}
	if !strings.Contains(mailer.url, "/auth/verify?token=") {
		t.Fatalf("expected verify URL to be generated, got %q", mailer.url)
	}

	verifyReq := httptest.NewRequest(http.MethodGet, mailer.url, nil)
	verifyReq.Host = "bearbeitung.example.org"
	verifyRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(verifyRec, verifyReq)

	if verifyRec.Code != http.StatusSeeOther {
		t.Fatalf("expected verify to redirect, got %d", verifyRec.Code)
	}
	cookies := verifyRec.Result().Cookies()
	if len(cookies) == 0 || cookies[0].Name != "static_editor_session" {
		t.Fatalf("expected session cookie to be set")
	}
}

func TestEditRequiresSessionAndMarksDocument(t *testing.T) {
	staticRoot := filepath.Join(t.TempDir(), "static")
	if err := os.MkdirAll(staticRoot, 0o755); err != nil {
		t.Fatalf("mkdir static root: %v", err)
	}
	if err := os.WriteFile(filepath.Join(staticRoot, "index.html"), []byte(`<!doctype html><html><body><main><h1>Hallo</h1><p>Welt</p></main></body></html>`), 0o644); err != nil {
		t.Fatalf("write html fixture: %v", err)
	}

	cfg := config.Config{
		Addr:          ":8090",
		DataDir:       t.TempDir(),
		SessionTTL:    "12h",
		MagicLinkTTL:  "15m",
		SecureCookies: false,
		Tenants: map[string]model.Tenant{
			"example.org": {
				Domain:            "example.org",
				LoginDomain:       "bearbeitung.example.org",
				AllowedEmails:     []string{"andy@example.org"},
				StartPath:         "/index.html",
				StaticRoot:        staticRoot,
				MainSelector:      "main",
				AllowedBlockTags:  []string{"h1", "p"},
				AllowedInlineTags: []string{"strong"},
			},
		},
	}

	app := New(cfg)
	mailer := &fakeMailer{}
	app.mailer = mailer

	loginReq := httptest.NewRequest(http.MethodPost, "http://bearbeitung.example.org/auth/request-link", strings.NewReader(url.Values{
		"email": []string{"andy@example.org"},
	}.Encode()))
	loginReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	loginReq.Header.Set("Host", "bearbeitung.example.org")
	loginRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(loginRec, loginReq)

	verifyReq := httptest.NewRequest(http.MethodGet, mailer.url, nil)
	verifyReq.Host = "bearbeitung.example.org"
	verifyRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(verifyRec, verifyReq)
	sessionCookie := verifyRec.Result().Cookies()[0]

	editReq := httptest.NewRequest(http.MethodGet, "http://bearbeitung.example.org/edit?path=/index.html", nil)
	editReq.Header.Set("Host", "bearbeitung.example.org")
	editReq.AddCookie(sessionCookie)
	editRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(editRec, editReq)

	if editRec.Code != http.StatusOK {
		t.Fatalf("expected edit route to return 200, got %d", editRec.Code)
	}
	body := editRec.Body.String()
	if !strings.Contains(body, `data-editor-id="node-0001"`) {
		t.Fatalf("expected prepared document markers in edit response")
	}
	if !strings.Contains(body, "Edit-Modus") {
		t.Fatalf("expected edit chrome in response")
	}
}
