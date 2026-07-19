package httpapp

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAdminUsersCRUD(t *testing.T) {
	app, sender, _ := setupAuthApp(t)
	sessionCookie := authenticateAdminSession(t, app, sender, "localhost:8080")

	createBody, _ := json.Marshal(map[string]any{
		"email":  "manager@example.com",
		"name":   "Manager",
		"role":   "event_manager",
		"status": "invited",
	})
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/users", bytes.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.AddCookie(sessionCookie)
	createRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(createRec, createReq)

	if createRec.Code != http.StatusCreated {
		t.Fatalf("expected create status 201, got %d", createRec.Code)
	}
	created := decodeBody[map[string]any](t, createRec)["item"].(map[string]any)
	userID := created["id"].(string)
	if created["email"] != "manager@example.com" {
		t.Fatalf("expected created email manager@example.com, got %v", created["email"])
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users", nil)
	listReq.AddCookie(sessionCookie)
	listRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("expected list status 200, got %d", listRec.Code)
	}
	listPayload := decodeBody[map[string]any](t, listRec)
	if listPayload["total"] != float64(2) {
		t.Fatalf("expected total 2, got %v", listPayload["total"])
	}

	patchBody, _ := json.Marshal(map[string]any{
		"role":   "admin",
		"status": "active",
	})
	patchReq := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/users/"+userID, bytes.NewReader(patchBody))
	patchReq.Header.Set("Content-Type", "application/json")
	patchReq.AddCookie(sessionCookie)
	patchRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(patchRec, patchReq)
	if patchRec.Code != http.StatusOK {
		t.Fatalf("expected patch status 200, got %d", patchRec.Code)
	}
	updated := decodeBody[map[string]any](t, patchRec)["item"].(map[string]any)
	if updated["role"] != "admin" {
		t.Fatalf("expected updated role admin, got %v", updated["role"])
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, "/api/v1/admin/users/"+userID, nil)
	deleteReq.AddCookie(sessionCookie)
	deleteRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(deleteRec, deleteReq)
	if deleteRec.Code != http.StatusNoContent {
		t.Fatalf("expected delete status 204, got %d", deleteRec.Code)
	}
}
