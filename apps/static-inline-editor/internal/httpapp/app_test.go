package httpapp

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
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
	if got := verifyRec.Result().Header.Get("Location"); got != "/edit?path=%2Findex.html" {
		t.Fatalf("expected verify redirect to edit start page, got %q", got)
	}
	cookies := verifyRec.Result().Cookies()
	if len(cookies) == 0 || cookies[0].Name != "static_editor_session" {
		t.Fatalf("expected session cookie to be set")
	}
}

func TestHomeLinksToEditStartPage(t *testing.T) {
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

	homeReq := httptest.NewRequest(http.MethodGet, "http://bearbeitung.example.org/", nil)
	homeReq.Host = "bearbeitung.example.org"
	homeReq.AddCookie(sessionCookie)
	homeRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(homeRec, homeReq)

	if homeRec.Code != http.StatusOK {
		t.Fatalf("expected home route to return 200, got %d", homeRec.Code)
	}
	if !strings.Contains(homeRec.Body.String(), `href="/edit?path=%2Findex.html"`) {
		t.Fatalf("expected home to link to edit start page, got %q", homeRec.Body.String())
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
	if strings.Contains(body, `class="frame"`) {
		t.Fatalf("expected original document to be edited in place, not embedded in wrapper frame")
	}
	if !strings.Contains(body, `static-inline-editor-bar`) {
		t.Fatalf("expected editor chrome to be injected into original document")
	}
}

func TestStaticAssetFallbackServesStylesheetFromTenantRoot(t *testing.T) {
	staticRoot := filepath.Join(t.TempDir(), "static")
	if err := os.MkdirAll(filepath.Join(staticRoot, "assets"), 0o755); err != nil {
		t.Fatalf("mkdir static root: %v", err)
	}
	if err := os.WriteFile(filepath.Join(staticRoot, "assets", "site.css"), []byte("body{color:red}"), 0o644); err != nil {
		t.Fatalf("write stylesheet fixture: %v", err)
	}

	cfg := config.Config{
		Addr:          ":8090",
		DataDir:       t.TempDir(),
		SessionTTL:    "12h",
		MagicLinkTTL:  "15m",
		SecureCookies: false,
		Tenants: map[string]model.Tenant{
			"example.org": {
				Domain:      "example.org",
				LoginDomain: "bearbeitung.example.org",
				StaticRoot:  staticRoot,
			},
		},
	}

	app := New(cfg)

	req := httptest.NewRequest(http.MethodGet, "http://bearbeitung.example.org/assets/site.css", nil)
	req.Host = "bearbeitung.example.org"
	rec := httptest.NewRecorder()
	app.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected stylesheet request to return 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "body{color:red}") {
		t.Fatalf("expected stylesheet content to be served")
	}
}

func TestPreviewAndSaveFlow(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	baseDir := t.TempDir()
	staticRoot := filepath.Join(baseDir, "static")
	backupRoot := filepath.Join(baseDir, "backups")
	repoRoot := filepath.Join(baseDir, "repo")
	remoteRoot := filepath.Join(baseDir, "remote.git")
	if err := os.MkdirAll(staticRoot, 0o755); err != nil {
		t.Fatalf("mkdir static root: %v", err)
	}
	if err := os.MkdirAll(backupRoot, 0o755); err != nil {
		t.Fatalf("mkdir backup root: %v", err)
	}
	if err := os.MkdirAll(repoRoot, 0o755); err != nil {
		t.Fatalf("mkdir repo root: %v", err)
	}
	if out, err := exec.Command("git", "init", "--bare", remoteRoot).CombinedOutput(); err != nil {
		t.Fatalf("git init --bare failed: %v: %s", err, out)
	}

	if out, err := exec.Command("git", "-C", repoRoot, "init").CombinedOutput(); err != nil {
		t.Fatalf("git init failed: %v: %s", err, out)
	}
	if out, err := exec.Command("git", "-C", repoRoot, "remote", "add", "origin", remoteRoot).CombinedOutput(); err != nil {
		t.Fatalf("git remote add failed: %v: %s", err, out)
	}

	targetFile := filepath.Join(repoRoot, "index.html")
	originalHTML := `<!doctype html><html><body><main><h1>Hallo</h1><p>Welt</p></main></body></html>`
	if err := os.WriteFile(targetFile, []byte(originalHTML), 0o644); err != nil {
		t.Fatalf("write html fixture: %v", err)
	}
	if out, err := exec.Command("git", "-C", repoRoot, "add", "--", "index.html").CombinedOutput(); err != nil {
		t.Fatalf("git add failed: %v: %s", err, out)
	}
	commit := exec.Command("git", "-C", repoRoot, "commit", "--message", "initial")
	commit.Env = append(commit.Environ(),
		"GIT_AUTHOR_NAME=Tester",
		"GIT_AUTHOR_EMAIL=tester@example.org",
		"GIT_COMMITTER_NAME=Tester",
		"GIT_COMMITTER_EMAIL=tester@example.org",
	)
	if out, err := commit.CombinedOutput(); err != nil {
		t.Fatalf("git commit failed: %v: %s", err, out)
	}

	cfg := config.Config{
		Addr:          ":8090",
		DataDir:       filepath.Join(baseDir, "data"),
		SessionTTL:    "12h",
		MagicLinkTTL:  "15m",
		SecureCookies: false,
		GitAuthorName: "Static Inline Editor",
		GitPushOnSave: true,
		GitRemoteName: "origin",
		Tenants: map[string]model.Tenant{
			"example.org": {
				Domain:            "example.org",
				LoginDomain:       "bearbeitung.example.org",
				AllowedEmails:     []string{"andy@example.org"},
				StartPath:         "/index.html",
				StaticRoot:        repoRoot,
				BackupRoot:        backupRoot,
				RepoRoot:          repoRoot,
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
	if strings.TrimSpace(saveResp.CommitHash) == "" {
		t.Fatalf("expected commit hash in save response")
	}
	if !saveResp.Pushed || strings.TrimSpace(saveResp.PushTarget) == "" {
		t.Fatalf("expected save response to report successful push")
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
	logOut, err := exec.Command("git", "-C", repoRoot, "log", "-1", "--pretty=%s").CombinedOutput()
	if err != nil {
		t.Fatalf("git log failed: %v: %s", err, logOut)
	}
	if !strings.Contains(string(logOut), "edit(example.org): /index.html by andy@example.org") {
		t.Fatalf("expected git commit message to mention edited file, got %q", string(logOut))
	}
	branchOut, err := exec.Command("git", "-C", repoRoot, "branch", "--show-current").CombinedOutput()
	if err != nil {
		t.Fatalf("git branch --show-current failed: %v: %s", err, branchOut)
	}
	branch := strings.TrimSpace(string(branchOut))
	remoteHashOut, err := exec.Command("git", "--git-dir", remoteRoot, "rev-parse", "--verify", "refs/heads/"+branch).CombinedOutput()
	if err != nil {
		t.Fatalf("remote rev-parse failed: %v: %s", err, remoteHashOut)
	}
	if strings.TrimSpace(string(remoteHashOut)) != saveResp.CommitHash {
		t.Fatalf("expected remote hash to match saved commit")
	}
}
