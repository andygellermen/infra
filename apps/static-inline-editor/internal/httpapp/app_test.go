package httpapp

import (
	"bytes"
	"encoding/json"
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
	if !strings.Contains(body, `data-editable`) {
		t.Fatalf("expected contenttools editable region marker in edit response")
	}
	if !strings.Contains(strings.ToLower(body), `content-tools.min.js`) {
		t.Fatalf("expected contenttools script to be injected on edit route")
	}
	if !strings.Contains(body, "Edit-Modus") {
		t.Fatalf("expected edit chrome in response")
	}
}

func TestPreviewAndSaveFlow(t *testing.T) {
	baseDir := t.TempDir()
	staticRoot := filepath.Join(baseDir, "static")
	backupRoot := filepath.Join(baseDir, "backups")
	if err := os.MkdirAll(staticRoot, 0o755); err != nil {
		t.Fatalf("mkdir static root: %v", err)
	}
	if err := os.MkdirAll(backupRoot, 0o755); err != nil {
		t.Fatalf("mkdir backup root: %v", err)
	}

	targetFile := filepath.Join(staticRoot, "index.html")
	originalHTML := `<!doctype html><html><body><main><h1>Hallo</h1><p>Welt</p></main></body></html>`
	if err := os.WriteFile(targetFile, []byte(originalHTML), 0o644); err != nil {
		t.Fatalf("write html fixture: %v", err)
	}

	cfg := config.Config{
		Addr:          ":8090",
		DataDir:       filepath.Join(baseDir, "data"),
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
				BackupRoot:        backupRoot,
				MainSelector:      "main",
				AllowedBlockTags:  []string{"h1", "p", "ul", "ol", "li"},
				AllowedInlineTags: []string{"strong", "em", "a", "br"},
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

	payload := model.PreviewRequest{
		Path: "/index.html",
		Regions: map[string]string{
			"main-content": `<h1>Neu</h1><p>Hallo <strong>Welt</strong></p>`,
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal preview payload: %v", err)
	}

	previewReq := httptest.NewRequest(http.MethodPost, "http://bearbeitung.example.org/preview", bytes.NewReader(body))
	previewReq.Header.Set("Content-Type", "application/json")
	previewReq.Header.Set("Host", "bearbeitung.example.org")
	previewReq.AddCookie(sessionCookie)
	previewRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(previewRec, previewReq)

	if previewRec.Code != http.StatusOK {
		t.Fatalf("expected preview to return 200, got %d", previewRec.Code)
	}
	var previewResp model.PreviewResponse
	if err := json.Unmarshal(previewRec.Body.Bytes(), &previewResp); err != nil {
		t.Fatalf("decode preview response: %v", err)
	}
	if !previewResp.OK || !strings.Contains(previewResp.PreviewHTML, "<h1>Neu</h1>") {
		t.Fatalf("expected preview html to contain updated content")
	}

	current, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("read target file: %v", err)
	}
	if string(current) != originalHTML {
		t.Fatalf("preview must not write file before save")
	}

	saveReq := httptest.NewRequest(http.MethodPost, "http://bearbeitung.example.org/save", bytes.NewReader(body))
	saveReq.Header.Set("Content-Type", "application/json")
	saveReq.Header.Set("Host", "bearbeitung.example.org")
	saveReq.AddCookie(sessionCookie)
	saveRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(saveRec, saveReq)

	if saveRec.Code != http.StatusOK {
		t.Fatalf("expected save to return 200, got %d", saveRec.Code)
	}
	var saveResp model.SaveResponse
	if err := json.Unmarshal(saveRec.Body.Bytes(), &saveResp); err != nil {
		t.Fatalf("decode save response: %v", err)
	}
	if !saveResp.OK || saveResp.BackupPath == "" {
		t.Fatalf("expected backup path in save response")
	}

	updated, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("read updated target file: %v", err)
	}
	if !strings.Contains(string(updated), "<h1>Neu</h1>") {
		t.Fatalf("expected saved file to contain updated content")
	}
	if _, err := os.Stat(saveResp.BackupPath); err != nil {
		t.Fatalf("expected backup file to exist: %v", err)
	}
}
