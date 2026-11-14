package secrets

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/SaranHiruthikM/envman/internal/middleware"
	"github.com/SaranHiruthikM/envman/internal/permissions"
	"github.com/SaranHiruthikM/envman/internal/response"
	"github.com/go-chi/chi/v5"
)

func (h *Handler) PushSecrets(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectID")
	if projectID == "" {
		response.Error(w, http.StatusBadRequest, "Missing URL params")
		return
	}

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	var req PushSecretsRequest
	if err := dec.Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Malformed request body")
		return
	}
	if req.Ciphertext == "" {
		response.Error(w, http.StatusBadRequest, "ciphertext is required")
		return
	}

	userID, ok := middleware.GetUserID(r)
	if !ok {
		response.Error(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	role, err := permissions.GetUserRole(r.Context(), h.db, userID, projectID)
	if err != nil {
		response.Error(w, http.StatusForbidden, "Not a member of this project")
		return
	}
	if strings.ToLower(role) != "owner" && strings.ToLower(role) != "coordinator" {
		response.Error(w, http.StatusForbidden, "Only owners or coordinators can push secrets")
		return
	}

	if err := h.pushSecrets(r.Context(), projectID, req.Ciphertext, userID); err != nil {
		response.Error(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	response.Json(w, http.StatusCreated, map[string]string{"message": "Pushed"})
}

func (h *Handler) GetSecrets(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectID")
	if projectID == "" {
		response.Error(w, http.StatusBadRequest, "Missing URL params")
		return
	}

	userID, ok := middleware.GetUserID(r)
	if !ok {
		response.Error(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	if _, err := permissions.GetUserRole(r.Context(), h.db, userID, projectID); err != nil {
		response.Error(w, http.StatusForbidden, "Not a member of this project")
		return
	}

	ciphertext, err := h.getSecrets(r.Context(), projectID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	response.Json(w, http.StatusOK, map[string]string{"ciphertext": ciphertext})
}
