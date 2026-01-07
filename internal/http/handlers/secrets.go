// Package handlers contains HTTP handlers for the GophKeeper server.
// Handlers are responsible for request parsing, response formatting,
// and delegating business logic to the service layer.
package handlers

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"gophkeeper/internal/http/middleware"
	"gophkeeper/internal/model"
	"gophkeeper/internal/repository"
	"gophkeeper/internal/service"
)

// SecretsHandler handles HTTP requests for managing user secrets.
// It provides endpoints for creating, retrieving, listing,
// and deleting encrypted secrets.
type SecretsHandler struct {
	svc *service.SecretsService
}

// NewSecretsHandler creates a new SecretsHandler using the provided SecretsService.
func NewSecretsHandler(svc *service.SecretsService) *SecretsHandler {
	return &SecretsHandler{svc: svc}
}

type upsertReq struct {
	ID            int64  `json:"id"`
	Type          string `json:"type"`
	Title         string `json:"title"`
	Meta          string `json:"meta"`
	EncryptedData string `json:"encrypted_data"` // base64
	Version       int64  `json:"version"`
	UpdatedAt     string `json:"updated_at"` // RFC3339
	Deleted       bool   `json:"deleted"`
}

// Upsert handles creation or update of a secret.
// It reads encrypted secret data from the request body
// and delegates persistence to the secrets service.
func (h *SecretsHandler) Upsert(w http.ResponseWriter, r *http.Request) {
	uid, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var req upsertReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	b, err := base64.StdEncoding.DecodeString(req.EncryptedData)
	if err != nil {
		http.Error(w, "bad encrypted_data", http.StatusBadRequest)
		return
	}
	tm := time.Now().UTC()
	if req.UpdatedAt != "" {
		if parsed, err := time.Parse(time.RFC3339Nano, req.UpdatedAt); err == nil {
			tm = parsed
		}
	}
	sec := &model.Secret{
		ID:            req.ID,
		UserID:        uid,
		Type:          model.SecretType(req.Type),
		Title:         req.Title,
		Meta:          req.Meta,
		EncryptedData: b,
		Version:       req.Version,
		UpdatedAt:     tm,
		Deleted:       req.Deleted,
	}
	out, err := h.svc.Upsert(r.Context(), sec)
	if err != nil {
		if err == repository.ErrNotFound {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"id": out.ID})
}

// List returns metadata for all secrets belonging to the authenticated user.
// Encrypted payloads are not included in the response.
func (h *SecretsHandler) List(w http.ResponseWriter, r *http.Request) {
	uid, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	items, err := h.svc.List(r.Context(), uid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	type item struct {
		ID        int64  `json:"id"`
		Type      string `json:"type"`
		Title     string `json:"title"`
		Meta      string `json:"meta"`
		Version   int64  `json:"version"`
		UpdatedAt string `json:"updated_at"`
		Deleted   bool   `json:"deleted"`
	}
	var resp []item
	for _, s := range items {
		resp = append(resp, item{
			ID: s.ID, Type: string(s.Type), Title: s.Title, Meta: s.Meta,
			Version: s.Version, UpdatedAt: s.UpdatedAt.Format(time.RFC3339Nano), Deleted: s.Deleted,
		})
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// Get retrieves a secret by its identifier.
// The handler returns encrypted data with server-side
// envelope encryption already removed.
func (h *SecretsHandler) Get(w http.ResponseWriter, r *http.Request) {
	uid, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	sec, err := h.svc.Get(r.Context(), id, uid)
	if err != nil {
		if err == repository.ErrNotFound {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	resp := upsertReq{
		ID:            sec.ID,
		Type:          string(sec.Type),
		Title:         sec.Title,
		Meta:          sec.Meta,
		EncryptedData: base64.StdEncoding.EncodeToString(sec.EncryptedData),
		Version:       sec.Version,
		UpdatedAt:     sec.UpdatedAt.Format(time.RFC3339Nano),
		Deleted:       sec.Deleted,
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// Delete removes a secret belonging to the authenticated user.
func (h *SecretsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	uid, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err := h.svc.Delete(r.Context(), id, uid); err != nil {
		if err == repository.ErrNotFound {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
