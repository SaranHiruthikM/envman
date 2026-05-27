package projects

import (
	"encoding/json"
	"net/http"

	"github.com/SaranHiruthikM/envman/internal/middleware"
	"github.com/SaranHiruthikM/envman/internal/permissions"
	"github.com/SaranHiruthikM/envman/internal/response"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
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

	response.Json(w, http.StatusCreated, CreateProjectResponse{ProjectID: projectId, Name: req.Name})
}

func (h *Handler) GetProjectByID(w http.ResponseWriter, r *http.Request) {

	projectID := chi.URLParam(r, "id")
	if projectID == "" {
		response.Error(w, http.StatusBadRequest, "Name is required")
		return
	}

	userID, ok := middleware.GetUserID(r)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Something went wrong")
		return
	}

	_, err := permissions.GetUserRole(r.Context(), h.db, userID, projectID)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "Not a part of requested project")
		return
	}

	project, err := h.getProjectByID(r.Context(), projectID)
	if err == pgx.ErrNoRows {
		response.Error(w, http.StatusNotFound, "Project not found")
		return
	}
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	response.Json(w, http.StatusOK, GetProjectResponse{Project: project})
}
