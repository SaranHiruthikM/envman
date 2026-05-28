package members

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/SaranHiruthikM/envman/internal/common"
	"github.com/SaranHiruthikM/envman/internal/middleware"
	"github.com/SaranHiruthikM/envman/internal/permissions"
	"github.com/SaranHiruthikM/envman/internal/response"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
)

func (h *Handler) InviteMembers(w http.ResponseWriter, r *http.Request) {
	var req InviteMembersRequest

	projectID := chi.URLParam(r, "id")
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Malformed Request Body")
		return
	}

	if req.Email == "" || req.Role == "" {
		response.Error(w, http.StatusBadRequest, "Missing Request fields")
		return
	}

	if strings.ToLower(req.Role) != "coordinator" && strings.ToLower(req.Role) != "member" {
		response.Error(w, http.StatusBadRequest, "Only can assign member or coordinator")
		return
	}

	userID, ok := middleware.GetUserID(r)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Something went wrong")
		return
	}

	role, err := permissions.GetUserRole(r.Context(), h.db, userID, projectID)
	if err != nil {
		response.Error(w, http.StatusForbidden, "Not a part of requested project")
		return
	}

	if strings.ToLower(role) != "owner" {
		response.Error(w, http.StatusForbidden, "Should be an owner to assign members")
		return
	}

	user, err := common.GetUserByEmail(r.Context(), h.db, req.Email)
	if err == pgx.ErrNoRows {
		response.Error(w, http.StatusNotFound, "User not found")
		return
	}

	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	_, err = permissions.GetUserRole(r.Context(), h.db, user.ID, projectID)
	if err == nil {
		response.Error(w, http.StatusConflict, "User is already a member")
		return
	}

	if err != pgx.ErrNoRows {
		response.Error(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	if err = h.inviteMembers(r.Context(), user.ID, projectID, req.Role); err != nil {
		response.Error(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	response.Json(w, http.StatusCreated, map[string]string{"message": "Successfully invited to team"})
}

func (h *Handler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	var req UpdateRoleRequest

	projectID := chi.URLParam(r, "id")
	targetUserID := chi.URLParam(r, "userId")
	if projectID == "" || targetUserID == "" {
		response.Error(w, http.StatusBadRequest, "Missing URL params")
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Malformed Request Body")
		return
	}

	if req.Role == "" {
		response.Error(w, http.StatusBadRequest, "Role is required")
		return
	}

	if strings.ToLower(req.Role) != "coordinator" && strings.ToLower(req.Role) != "member" {
		response.Error(w, http.StatusBadRequest, "Only can assign member or coordinator")
		return
	}

	requestingUserID, ok := middleware.GetUserID(r)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Something went wrong")
		return
	}

	if requestingUserID == targetUserID {
		response.Error(w, http.StatusBadRequest, "Owner cannot change their own role")
		return
	}

	role, err := permissions.GetUserRole(r.Context(), h.db, requestingUserID, projectID)
	if err != nil {
		response.Error(w, http.StatusForbidden, "Not a part of requested project")
		return
	}

	if strings.ToLower(role) != "owner" {
		response.Error(w, http.StatusForbidden, "Should be an owner to change roles")
		return
	}

	_, err = permissions.GetUserRole(r.Context(), h.db, targetUserID, projectID)
	if err == pgx.ErrNoRows {
		response.Error(w, http.StatusNotFound, "Target user is not a member")
		return
	}

	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	if err = h.updateRole(r.Context(), targetUserID, projectID, req.Role); err != nil {
		response.Error(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	response.Json(w, http.StatusOK, map[string]string{"message": "Successfully updated role"})
}

func (h *Handler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "id")
	targetUserID := chi.URLParam(r, "userId")
	if projectID == "" || targetUserID == "" {
		response.Error(w, http.StatusBadRequest, "Missing URL params")
		return
	}

	requestingUserID, ok := middleware.GetUserID(r)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Something went wrong")
		return
	}

	role, err := permissions.GetUserRole(r.Context(), h.db, requestingUserID, projectID)
	if err != nil {
		response.Error(w, http.StatusForbidden, "Not a part of requested project")
		return
	}

	if strings.ToLower(role) != "owner" {
		response.Error(w, http.StatusForbidden, "Should be an owner to remove members")
		return
	}

	if requestingUserID == targetUserID {
		count, err := h.countOwners(r.Context(), projectID)
		if err != nil {
			response.Error(w, http.StatusInternalServerError, "Something went wrong")
			return
		}
		if count <= 1 {
			response.Error(w, http.StatusBadRequest, "Cannot remove yourself — you are the only owner")
			return
		}
	}

	_, err = permissions.GetUserRole(r.Context(), h.db, targetUserID, projectID)
	if err == pgx.ErrNoRows {
		response.Error(w, http.StatusNotFound, "Target user is not a member")
		return
	}

	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	if err = h.removeMember(r.Context(), targetUserID, projectID); err != nil {
		response.Error(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	response.Json(w, http.StatusOK, map[string]string{"message": "Successfully removed member"})
}
