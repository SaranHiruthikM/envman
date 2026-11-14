package audit

import (
	"net/http"

	"github.com/SaranHiruthikM/envman/internal/middleware"
	"github.com/SaranHiruthikM/envman/internal/permissions"
	"github.com/SaranHiruthikM/envman/internal/response"
	"github.com/go-chi/chi/v5"
)

func (h *Handler) GetAuditLogs(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectID")
	if projectID == "" {
		response.Error(w, http.StatusBadRequest, "projectID is required")
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

	logs, err := h.getLogs(r.Context(), projectID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	response.Json(w, http.StatusOK, logs)
}
