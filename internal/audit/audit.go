package audit

import (
	"net/http"

	"github.com/SaranHiruthikM/envman/internal/middleware"
	"github.com/SaranHiruthikM/envman/internal/models"
	"github.com/SaranHiruthikM/envman/internal/permissions"
	"github.com/SaranHiruthikM/envman/internal/response"
	"github.com/go-chi/chi/v5"
)

func (h *Handler) GetAuditLogs(w http.ResponseWriter, r *http.Request) {
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

	logs, err := h.getLogs(r.Context(), projectID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	response.Json(w, http.StatusOK, map[string][]models.Logs{"logs": logs})
}
