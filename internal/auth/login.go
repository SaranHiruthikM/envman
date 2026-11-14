package auth

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"sync"

	"github.com/SaranHiruthikM/envman/internal/common"
	"github.com/SaranHiruthikM/envman/internal/response"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

// dummyHash is a real bcrypt hash of random bytes. We run bcrypt against it
// when the email is unknown, so the wall-clock time matches the real path and
// no timing oracle reveals whether an email is registered.
var (
	dummyHashOnce sync.Once
	dummyHash     []byte
)

func getDummyHash() []byte {
	dummyHashOnce.Do(func() {
		buf := make([]byte, 32)
		_, _ = rand.Read(buf)
		h, err := bcrypt.GenerateFromPassword(buf, bcrypt.DefaultCost)
		if err != nil {
			// Fall back to any valid bcrypt hash; the string below is a known-good
			// hash of a random secret and will never collide with a user password.
			h = []byte("$2a$10$abcdefghijklmnopqrstuvwxyz0123456789ABCDEFGHIJKLMNOPQR")
		}
		dummyHash = h
	})
	return dummyHash
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	var req LoginRequest
	if err := dec.Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Malformed request body")
		return
	}

	if req.Email == "" || req.Password == "" {
		response.Error(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	user, err := common.GetUserByEmail(r.Context(), h.db, req.Email)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		response.Error(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	hash := getDummyHash()
	if user != nil {
		hash = []byte(user.PasswordHash)
	}
	bcryptErr := bcrypt.CompareHashAndPassword(hash, []byte(req.Password))

	if user == nil || bcryptErr != nil {
		response.Error(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	token, err := h.tokenManager.Generate(user.ID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	response.Json(w, http.StatusOK, LoginResponse{
		Token:                    token,
		UserID:                   user.ID,
		PublicKey:                base64.StdEncoding.EncodeToString(user.PublicKey),
		EncryptedPrivateKey:      base64.StdEncoding.EncodeToString(user.EncryptedPrivateKey),
		EncryptedPrivateKeyNonce: base64.StdEncoding.EncodeToString(user.EncryptedPrivateKeyNonce),
		KDFSalt:                  base64.StdEncoding.EncodeToString(user.KDFSalt),
		KDFTime:                  user.KDFTime,
		KDFMemory:                user.KDFMemory,
		KDFParallelism:           user.KDFParallelism,
	})
}
