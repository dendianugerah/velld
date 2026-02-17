package auth

import (
	"encoding/json"
	"net/http"

	"github.com/dendianugerah/velld/internal/common"
	"github.com/dendianugerah/velld/internal/common/response"
)

type AuthHandler struct {
	authService *AuthService
}

func NewAuthHandler(authService *AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if !common.GetSecrets().IsAllowSignup {
		response.SendError(w, http.StatusForbidden, "Registration is disabled")
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.SendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.authService.Register(req.Username, req.Password); err != nil {
		response.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.SendSuccess(w, "Registration successful", ProfileResponse{Username: req.Username})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.SendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	token, err := h.authService.Login(req.Username, req.Password)
	if err != nil {
		response.SendError(w, http.StatusUnauthorized, err.Error())
		return
	}

	response.SendSuccess(w, "Login successful", LoginResponse{Token: token})
}

func (h *AuthHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	username, err := h.authService.GetProfile(r.Context())
	if err != nil {
		response.SendError(w, http.StatusUnauthorized, err.Error())
		return
	}

	response.SendSuccess(w, "Profile retrieved successfully", ProfileResponse{Username: username})
}