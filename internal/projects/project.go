package projects

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/SaranHiruthikM/envman/internal/middleware"
	"github.com/SaranHiruthikM/envman/internal/permissions"
	"github.com/SaranHiruthikM/envman/internal/response"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
)

const maxProjectNameLen = 120

func (h *Handler) CreateProject(w http.ResponseWriter, r *http.Request) {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	var req ProjectRequest
	if err := dec.Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Malformed request body")
		return
	}
	if req.Name == "" {
		response.Error(w, http.StatusBadRequest, "Name is required")
		return
	}
	if len(req.Name) > maxProjectNameLen {
		response.Error(w, http.StatusBadRequest, "Name is too long")
		return
	}

	userID, ok := middleware.GetUserID(r)
	if !ok {
		response.Error(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	projectID, err := h.createProject(r.Context(), userID, req.Name)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	response.Json(w, http.StatusCreated, CreateProjectResponse{ProjectID: projectID, Name: req.Name})
}

func (h *Handler) GetProjectByID(w http.ResponseWriter, r *http.Request) {
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
		if errors.Is(err, pgx.ErrNoRows) {
			response.Error(w, http.StatusForbidden, "Not a member of this project")
			return
		}
		response.Error(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	project, err := h.getProjectByID(r.Context(), projectID)
	if errors.Is(err, pgx.ErrNoRows) {
		response.Error(w, http.StatusNotFound, "Project not found")
		return
	}
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	response.Json(w, http.StatusOK, project)
}
