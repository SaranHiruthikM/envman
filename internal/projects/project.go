package projects

import (
	"encoding/json"
	"net/http"

	"github.com/SaranHiruthikM/envman/internal/middleware"
	"github.com/SaranHiruthikM/envman/internal/response"
)

func (h *Handler) CreateProject(w http.ResponseWriter, r *http.Request) {
	var req ProjectRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Malformed Request Body")
		return
	}

	if req.Name == "" {
		response.Error(w, http.StatusBadRequest, "Name is required")
		return
	}

	userID, ok := middleware.GetUserID(r)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Something went wrong")
		return
	}

	projectId, err := h.createProject(r.Context(), userID, req.Name)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	response.Json(w, http.StatusCreated, ProjectResponse{ProjectID: projectId, Name: req.Name})
}
