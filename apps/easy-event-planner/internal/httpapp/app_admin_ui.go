package httpapp

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed adminui/*
var adminUIEmbeddedFiles embed.FS

var adminUIFiles = func() fs.FS {
	sub, err := fs.Sub(adminUIEmbeddedFiles, "adminui")
	if err != nil {
		panic(err)
	}
	return sub
}()

func (a *App) handleAdminUIRoutes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimSpace(r.URL.Path)
	switch path {
	case "/admin":
		a.serveAdminUIAsset(w, r, "index.html", "text/html; charset=utf-8", true)
		return
	case "/admin/":
		http.Redirect(w, r, "/admin", http.StatusTemporaryRedirect)
		return
	case "/admin-ui.css":
		a.serveAdminUIAsset(w, r, "app.css", "text/css; charset=utf-8", false)
		return
	case "/admin-ui.js":
		a.serveAdminUIAsset(w, r, "app.js", "application/javascript; charset=utf-8", false)
		return
	default:
		if strings.HasPrefix(path, "/admin/") {
			// Keep tenant slug "admin" assets reachable via /admin/include.js etc.
			a.handleTenantAssetRoutes(w, r)
			return
		}
		http.NotFound(w, r)
		return
	}
}

func (a *App) serveAdminUIAsset(w http.ResponseWriter, _ *http.Request, name, contentType string, noStore bool) {
	payload, err := fs.ReadFile(adminUIFiles, name)
	if err != nil {
		http.Error(w, "admin ui asset not found", http.StatusNotFound)
		return
	}

	if noStore {
		w.Header().Set("Cache-Control", "no-store")
	}
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(payload)
}
