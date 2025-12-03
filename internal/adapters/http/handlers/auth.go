package handlers

import (
	"net/http"

	"github.com/mohits-git/watch-expense/internal/adapters/http/dtos"
	"github.com/mohits-git/watch-expense/internal/services"
	"github.com/mohits-git/watch-expense/internal/utils/apperr"
	"github.com/mohits-git/watch-expense/internal/utils/authctx"
)

type AuthHandler struct {
	authService services.AuthenticationService
}

func NewAuthHandler(authService services.AuthenticationService) *AuthHandler {
	return &AuthHandler{authService}
}

func (h *AuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	loginRequest, err := decodeRequest[dtos.LoginRequest](r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	token, err := h.authService.Login(r.Context(), loginRequest.Email, loginRequest.Password)
	if err != nil {
		if apperr.IsNotFoundError(err) {
			writeError(w, http.StatusUnauthorized, "invalid email or password")
		} else if apperr.IsUnauthorizedError(err) {
			writeError(w, http.StatusUnauthorized, "invalid email or password")
		} else if apperr.IsInvalidError(err) {
			writeError(w, http.StatusBadRequest, "invalid inputs")
		} else {
			writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	loginResponse := dtos.LoginResponse{Token: token}
	w.Header().Set("Set-Cookie", "token="+token+"; HttpOnly; Path=/api/; SameSite=Strict")
	writeResponse(w, http.StatusOK, "login successful", loginResponse)
}

func (h *AuthHandler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	token, ok := authctx.TokenFromCtx(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	err := h.authService.Logout(r.Context(), token)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	logoutResponse := dtos.LogoutResponse{}
	w.Header().Set("Set-Cookie", "token=; HttpOnly; Path=/api/; Max-Age=0; SameSite=Strict")
	writeResponse(w, http.StatusOK, "logout successful", logoutResponse)
}

func (h *AuthHandler) HandleMe(w http.ResponseWriter, r *http.Request) {
	user, err := h.authService.GetCurrentUser(r.Context())
	if err != nil {
		if apperr.IsUnauthorizedError(err) {
			writeError(w, http.StatusUnauthorized, "unauthorized")
		} else if apperr.IsNotFoundError(err) {
			writeError(w, http.StatusNotFound, "user not found")
		} else {
			writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	userResponse := dtos.ToUserDTO(user)
	writeResponse(w, http.StatusOK, "user details fetched successfully", userResponse)
}
