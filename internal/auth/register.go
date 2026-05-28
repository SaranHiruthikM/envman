package auth

import (
	"encoding/json"
	"net/http"

	"github.com/SaranHiruthikM/envman/internal/common"
	"github.com/SaranHiruthikM/envman/internal/response"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Malformed Request Body")
		return
	}

	if req.Email == "" || req.Name == "" || req.Password == "" {
		response.Error(w, http.StatusBadRequest, "Malformed Request Body")
		return
	}

	user, err := common.GetUserByEmail(r.Context(), h.db, req.Email)
	if err != nil && err != pgx.ErrNoRows {
		response.Error(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	if user != nil {
		response.Error(w, http.StatusConflict, "User already exists")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	userId, err := h.createUser(r.Context(), req.Name, req.Email, string(hashedPassword))
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	token, err := h.tokenManager.Generate(userId)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	response.Json(w, http.StatusCreated, AuthResponse{Token: token})

}
