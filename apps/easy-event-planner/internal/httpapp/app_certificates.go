package httpapp

import (
	"net/http"
	"strings"
	"time"

	"github.com/andygellermann/infra/apps/easy-event-planner/internal/tenant"
)

func (a *App) handlePublicParticipantPortalCertificates(w http.ResponseWriter, r *http.Request, tenantItem tenant.Tenant) {
	if r.Method != http.MethodGet {
		writeAPIError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Methode nicht erlaubt.")
		return
	}
	if a.certificateService == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Zertifikats-Service ist nicht verfuegbar.")
		return
	}

	principal, ok := a.requireParticipantPrincipal(w, r, tenantItem)
	if !ok {
		return
	}

	items, err := a.certificateService.ListParticipantCertificates(r.Context(), tenantItem.ID, principal.ParticipantID)
	if err != nil {
		a.writeCertificateError(w, err)
		return
	}

	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		payload := certificatePayload(item)
		payload["download_url"] = "/api/v1/public/" + tenantItem.Slug + "/participants/portal/certificates/" + item.ID + "/download"
		result = append(result, payload)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"items": result,
		"total": len(result),
	})
}

func (a *App) handlePublicParticipantPortalCertificateGet(w http.ResponseWriter, r *http.Request, tenantItem tenant.Tenant, certificateID string) {
	if r.Method != http.MethodGet {
		writeAPIError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Methode nicht erlaubt.")
		return
	}
	if a.certificateService == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Zertifikats-Service ist nicht verfuegbar.")
		return
	}

	principal, ok := a.requireParticipantPrincipal(w, r, tenantItem)
	if !ok {
		return
	}

	item, err := a.certificateService.GetParticipantCertificate(
		r.Context(),
		tenantItem.ID,
		principal.ParticipantID,
		certificateID,
	)
	if err != nil {
		a.writeCertificateError(w, err)
		return
	}

	payload := certificatePayload(item)
	payload["download_url"] = "/api/v1/public/" + tenantItem.Slug + "/participants/portal/certificates/" + item.ID + "/download"
	writeJSON(w, http.StatusOK, map[string]any{
		"item": payload,
	})
}

func (a *App) handlePublicParticipantPortalCertificateDownload(w http.ResponseWriter, r *http.Request, tenantItem tenant.Tenant, certificateID string) {
	if r.Method != http.MethodGet {
		writeAPIError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Methode nicht erlaubt.")
		return
	}
	if a.certificateService == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Zertifikats-Service ist nicht verfuegbar.")
		return
	}

	principal, ok := a.requireParticipantPrincipal(w, r, tenantItem)
	if !ok {
		return
	}

	item, bytes, err := a.certificateService.LoadParticipantCertificatePDF(
		r.Context(),
		tenantItem.ID,
		principal.ParticipantID,
		certificateID,
	)
	if err != nil {
		a.writeCertificateError(w, err)
		return
	}

	fileName := strings.ToLower(strings.TrimSpace(item.CertificateNumber))
	if fileName == "" {
		fileName = "certificate"
	}
	fileName = strings.ReplaceAll(fileName, " ", "-") + ".pdf"

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Cache-Control", "private, max-age=0, no-store")
	w.Header().Set("Content-Disposition", `inline; filename="`+fileName+`"`)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(bytes)
}

func (a *App) handlePublicCertificateVerify(w http.ResponseWriter, r *http.Request, tenantItem tenant.Tenant) {
	if r.Method != http.MethodGet {
		writeAPIError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Methode nicht erlaubt.")
		return
	}
	if a.certificateService == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Zertifikats-Service ist nicht verfuegbar.")
		return
	}

	certificateNumber := strings.TrimSpace(r.URL.Query().Get("certificate_no"))
	code := strings.TrimSpace(r.URL.Query().Get("code"))
	result, err := a.certificateService.VerifyByNumberAndCode(r.Context(), tenantItem.ID, certificateNumber, code)
	if err != nil {
		a.writeCertificateError(w, err)
		return
	}

	payload := certificatePayload(result.Certificate)
	writeJSON(w, http.StatusOK, map[string]any{
		"valid":       true,
		"verified_at": result.VerifiedAt.UTC().Format(time.RFC3339),
		"item":        payload,
	})
}
