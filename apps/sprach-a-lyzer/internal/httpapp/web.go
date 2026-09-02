package httpapp

import (
	"embed"
	"net/http"
	"path"
)

//go:embed web/*
var productFiles embed.FS

func (a *App) productUI(response http.ResponseWriter, request *http.Request) {
	serveProductFile(response, "web/index.html", "text/html; charset=utf-8")
}

func (a *App) adminUI(response http.ResponseWriter, request *http.Request) {
	serveProductFile(response, "web/admin.html", "text/html; charset=utf-8")
}

func (a *App) productAsset(response http.ResponseWriter, request *http.Request) {
	name := path.Base(request.URL.Path)
	contentType := "text/plain; charset=utf-8"
	if name == "app.css" {
		contentType = "text/css; charset=utf-8"
	} else if name == "app.js" {
		contentType = "text/javascript; charset=utf-8"
	}
	serveProductFile(response, "web/"+name, contentType)
}

func serveProductFile(response http.ResponseWriter, name, contentType string) {
	data, err := productFiles.ReadFile(name)
	if err != nil {
		http.Error(response, "not found", http.StatusNotFound)
		return
	}
	response.Header().Set("Content-Type", contentType)
	response.WriteHeader(http.StatusOK)
	_, _ = response.Write(data)
}
