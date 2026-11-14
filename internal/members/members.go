package members

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/SaranHiruthikM/envman/internal/common"
	"github.com/SaranHiruthikM/envman/internal/middleware"
	"github.com/SaranHiruthikM/envman/internal/permissions"
	"github.com/SaranHiruthikM/envman/internal/response"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
)

func decodeWrapFields(wrappedKey, nonce, ephPub string) (wrappedKey_b []byte, nonce_b []byte, ephPub_b []byte, err error) {
	wrappedKey_b, err = base64.StdEncoding.DecodeString(wrappedKey)
	if err != nil || len(wrappedKey_b) == 0 {
		return nil, nil, nil, errors.New("invalid wrapped_key")
	}
	nonce_b, err = base64.StdEncoding.DecodeString(nonce)
	if err != nil || len(nonce_b) != 24 {
		return nil, nil, nil, errors.New("invalid wrapped_key_nonce")
	}
	ephPub_b, err = base64.StdEncoding.DecodeString(ephPub)
	if err != nil || len(ephPub_b) != 32 {
		return nil, nil, nil, errors.New("invalid ephemeral_pubkey")
	}
	return wrappedKey_b, nonce_b, ephPub_b, nil
}

func (h *Handler) InviteMembers(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectID")

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	var req InviteMembersRequest
	if err := dec.Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Malformed request body")
		return
	}
	if req.Email == "" || req.Role == "" {
		response.Error(w, http.StatusBadRequest, "Email and role are required")
		return
	}
	role := strings.ToLower(req.Role)
	if role != "coordinator" && role != "member" {
		response.Error(w, http.StatusBadRequest, "Role must be 'member' or 'coordinator'")
		return
	}

	wk, nonce, ephPub, err := decodeWrapFields(req.WrappedKey, req.WrappedKeyNonce, req.EphemeralPubkey)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	requester, ok := middleware.GetUserID(r)
	if !ok {
		response.Error(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	requesterRole, err := permissions.GetUserRole(r.Context(), h.db, requester, projectID)
	if err != nil {
		response.Error(w, http.StatusForbidden, "Not a member of this project")
		return
	}
	if strings.ToLower(requesterRole) != "owner" {
		response.Error(w, http.StatusForbidden, "Only owners can invite")
		return
	}

	user, err := common.GetUserByEmail(r.Context(), h.db, req.Email)
	if errors.Is(err, pgx.ErrNoRows) {
		response.Error(w, http.StatusNotFound, "User not found")
		return
	}
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	if err := h.inviteMember(r.Context(), user.ID, projectID, role, wrappedKey{
		WrappedKey: wk, WrappedKeyNonce: nonce, EphemeralPubkey: ephPub,
	}); err != nil {
		if errors.Is(err, ErrAlreadyMember) {
			response.Error(w, http.StatusConflict, "User is already a member")
			return
		}
		response.Error(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	response.Json(w, http.StatusCreated, map[string]string{"message": "Invited"})
}

func (h *Handler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectID")
	targetUserID := chi.URLParam(r, "memberUserID")
	if projectID == "" || targetUserID == "" {
		response.Error(w, http.StatusBadRequest, "Missing URL params")
		return
	}

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	var req UpdateRoleRequest
	if err := dec.Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Malformed request body")
		return
	}
	role := strings.ToLower(req.Role)
	if role != "coordinator" && role != "member" {
		response.Error(w, http.StatusBadRequest, "Role must be 'member' or 'coordinator'")
		return
	}

	requester, ok := middleware.GetUserID(r)
	if !ok {
		response.Error(w, http.StatusInternalServerError, "Something went wrong")
		return
	}
	if requester == targetUserID {
		response.Error(w, http.StatusBadRequest, "Cannot change your own role")
		return
	}

	requesterRole, err := permissions.GetUserRole(r.Context(), h.db, requester, projectID)
	if err != nil {
		response.Error(w, http.StatusForbidden, "Not a member of this project")
		return
	}
	if strings.ToLower(requesterRole) != "owner" {
		response.Error(w, http.StatusForbidden, "Only owners can change roles")
		return
	}

	targetRole, err := permissions.GetUserRole(r.Context(), h.db, targetUserID, projectID)
	if errors.Is(err, pgx.ErrNoRows) {
		response.Error(w, http.StatusNotFound, "Target user is not a member")
		return
	}
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	// Demoting the last owner would leave the project ownerless.
	if strings.ToLower(targetRole) == "owner" {
		count, err := h.countOwners(r.Context(), projectID)
		if err != nil {
			response.Error(w, http.StatusInternalServerError, "Something went wrong")
			return
		}
		if count <= 1 {
			response.Error(w, http.StatusBadRequest, "Cannot demote the only owner")
			return
		}
	}

	if err := h.updateRole(r.Context(), targetUserID, projectID, role); err != nil {
		response.Error(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	response.Json(w, http.StatusOK, map[string]string{"message": "Role updated"})
}

func (h *Handler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectID")
	targetUserID := chi.URLParam(r, "memberUserID")
	if projectID == "" || targetUserID == "" {
		response.Error(w, http.StatusBadRequest, "Missing URL params")
		return
	}

	requester, ok := middleware.GetUserID(r)
	if !ok {
		response.Error(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	requesterRole, err := permissions.GetUserRole(r.Context(), h.db, requester, projectID)
	if err != nil {
		response.Error(w, http.StatusForbidden, "Not a member of this project")
		return
	}
	if strings.ToLower(requesterRole) != "owner" {
		response.Error(w, http.StatusForbidden, "Only owners can remove members")
		return
	}

	targetRole, err := permissions.GetUserRole(r.Context(), h.db, targetUserID, projectID)
	if errors.Is(err, pgx.ErrNoRows) {
		response.Error(w, http.StatusNotFound, "Target user is not a member")
		return
	}
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	if strings.ToLower(targetRole) == "owner" {
		count, err := h.countOwners(r.Context(), projectID)
		if err != nil {
			response.Error(w, http.StatusInternalServerError, "Something went wrong")
			return
		}
		if count <= 1 {
			response.Error(w, http.StatusBadRequest, "Cannot remove the only owner")
			return
		}
	}

	if err := h.removeMember(r.Context(), targetUserID, projectID); err != nil {
		response.Error(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	response.Json(w, http.StatusOK, map[string]string{"message": "Removed"})
}

// SetSelfKey stores the caller's own wrapped project key, used by `init` right
// after a project is created.
func (h *Handler) SetSelfKey(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectID")

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	var req SetSelfKeyRequest
	if err := dec.Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Malformed request body")
		return
	}

	wk, nonce, ephPub, err := decodeWrapFields(req.WrappedKey, req.WrappedKeyNonce, req.EphemeralPubkey)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
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

	if err := h.setSelfKey(r.Context(), userID, projectID, wrappedKey{
		WrappedKey: wk, WrappedKeyNonce: nonce, EphemeralPubkey: ephPub,
	}); err != nil {
		response.Error(w, http.StatusInternalServerError, "Something went wrong")
		return
	}
	response.Json(w, http.StatusCreated, map[string]string{"message": "Key stored"})
}

// GetSelfKey returns the caller's own wrapped project key.
func (h *Handler) GetSelfKey(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectID")

	userID, ok := middleware.GetUserID(r)
	if !ok {
		response.Error(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	if _, err := permissions.GetUserRole(r.Context(), h.db, userID, projectID); err != nil {
		response.Error(w, http.StatusForbidden, "Not a member of this project")
		return
	}

	k, err := h.getSelfKey(r.Context(), userID, projectID)
	if errors.Is(err, pgx.ErrNoRows) {
		response.Error(w, http.StatusNotFound, "No project key for this user")
		return
	}
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	response.Json(w, http.StatusOK, map[string]string{
		"wrapped_key":       base64.StdEncoding.EncodeToString(k.WrappedKey),
		"wrapped_key_nonce": base64.StdEncoding.EncodeToString(k.WrappedKeyNonce),
		"ephemeral_pubkey":  base64.StdEncoding.EncodeToString(k.EphemeralPubkey),
	})
}
