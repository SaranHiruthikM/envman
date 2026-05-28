package auth

import (
	"encoding/json"
	"net/http"

	"github.com/SaranHiruthikM/envman/internal/common"
	"github.com/SaranHiruthikM/envman/internal/response"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Malformed Request Body")
		return
	}

	if req.Email == "" || req.Password == "" {
		response.Error(w, http.StatusBadRequest, "Malformed Request Body")
		return
	}

	user, err := common.GetUserByEmail(r.Context(), h.db, req.Email)
	if err != nil && err != pgx.ErrNoRows {
		response.Error(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	if user == nil {
		response.Error(w, http.StatusUnauthorized, "Invalid Credentials")
		return
	}

	if err := checkCredentials(user.PasswordHash, req.Password); err != nil {
		response.Error(w, http.StatusUnauthorized, "Invalid Credentials")
		return
	}

	token, err := h.tokenManager.Generate(user.ID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	response.Json(w, http.StatusOK, AuthResponse{Token: token})

}

func checkCredentials(hashedPassword, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return err
	}

	return nil
}
