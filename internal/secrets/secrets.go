package secrets

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/SaranHiruthikM/envman/internal/middleware"
	"github.com/SaranHiruthikM/envman/internal/permissions"
	"github.com/SaranHiruthikM/envman/internal/response"
	"github.com/go-chi/chi/v5"
)

func (h *Handler) PushSecrets(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "id")
	if projectID == "" {
		response.Error(w, http.StatusBadRequest, "Missing URL params")
		return
	}

	var req PushSecretsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Malformed Request Body")
		return
	}

	userID, ok := middleware.GetUserID(r)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Something went wrong")
		return
	}

	role, err := permissions.GetUserRole(r.Context(), h.db, userID, projectID)
	log.Printf("role: %s, err: %v", role, err)
	if err != nil {
		response.Error(w, http.StatusForbidden, "Not a part of requested project")
		return
	}

	if strings.ToLower(role) != "owner" && strings.ToLower(role) != "coordinator" {
		response.Error(w, http.StatusForbidden, "Should be an owner or coordinator to push secrets")
		return
	}

	err = h.pushSecrets(r.Context(), projectID, req.EncryptedBlob, userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	response.Json(w, http.StatusCreated, map[string]string{"message": "Successfully pushed secrets"})
}

func (h *Handler) GetSecrets(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "id")
	if projectID == "" {
		response.Error(w, http.StatusBadRequest, "Missing URL params")
		return
	}

	userID, ok := middleware.GetUserID(r)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Something went wrong")
		return
	}

	_, err := permissions.GetUserRole(r.Context(), h.db, userID, projectID)
	if err != nil {
		response.Error(w, http.StatusForbidden, "Not a part of requested project")
		return
	}

	secret, err := h.getSecrets(r.Context(), projectID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	response.Json(w, http.StatusOK, map[string]string{"message": secret})
}
