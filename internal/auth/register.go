package auth

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/mail"

	"github.com/SaranHiruthikM/envman/internal/common"
	"github.com/SaranHiruthikM/envman/internal/response"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

const (
	minPasswordLen = 8
	maxPasswordLen = 72 // bcrypt silently truncates past 72 bytes
	maxNameLen     = 120
)

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	var req RegisterRequest
	if err := dec.Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Malformed request body")
		return
	}

	if req.Email == "" || req.Name == "" || req.Password == "" {
		response.Error(w, http.StatusBadRequest, "Email, name, and password are required")
		return
	}
	if _, err := mail.ParseAddress(req.Email); err != nil {
		response.Error(w, http.StatusBadRequest, "Email is not a valid address")
		return
	}
	if len(req.Password) < minPasswordLen || len(req.Password) > maxPasswordLen {
		response.Error(w, http.StatusBadRequest, "Password must be 8–72 characters")
		return
	}
	if len(req.Name) > maxNameLen {
		response.Error(w, http.StatusBadRequest, "Name is too long")
		return
	}
	if req.PublicKey == "" || req.EncryptedPrivateKey == "" ||
		req.EncryptedPrivateKeyNonce == "" || req.KDFSalt == "" ||
		req.KDFTime == 0 || req.KDFMemory == 0 || req.KDFParallelism == 0 {
		response.Error(w, http.StatusBadRequest, "Key material is required")
		return
	}

	pub, err := base64.StdEncoding.DecodeString(req.PublicKey)
	if err != nil || len(pub) != 32 {
		response.Error(w, http.StatusBadRequest, "Invalid public_key")
		return
	}
	encPriv, err := base64.StdEncoding.DecodeString(req.EncryptedPrivateKey)
	if err != nil || len(encPriv) == 0 {
		response.Error(w, http.StatusBadRequest, "Invalid encrypted_private_key")
		return
	}
	encPrivNonce, err := base64.StdEncoding.DecodeString(req.EncryptedPrivateKeyNonce)
	if err != nil || len(encPrivNonce) != 24 {
		response.Error(w, http.StatusBadRequest, "Invalid encrypted_private_key_nonce")
		return
	}
	salt, err := base64.StdEncoding.DecodeString(req.KDFSalt)
	if err != nil || len(salt) < 8 {
		response.Error(w, http.StatusBadRequest, "Invalid kdf_salt")
		return
	}

	existing, err := common.GetUserByEmail(r.Context(), h.db, req.Email)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		response.Error(w, http.StatusInternalServerError, "Something went wrong")
		return
	}
	if existing != nil {
		response.Error(w, http.StatusConflict, "User already exists")
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	userID, err := h.createUser(r.Context(), createUserParams{
		Email:                    req.Email,
		Name:                     req.Name,
		PasswordHash:             string(hashed),
		PublicKey:                pub,
		EncryptedPrivateKey:      encPriv,
		EncryptedPrivateKeyNonce: encPrivNonce,
		KDFSalt:                  salt,
		KDFTime:                  req.KDFTime,
		KDFMemory:                req.KDFMemory,
		KDFParallelism:           req.KDFParallelism,
	})
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	token, err := h.tokenManager.Generate(userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	response.Json(w, http.StatusCreated, RegisterResponse{Token: token, UserID: userID})
}
