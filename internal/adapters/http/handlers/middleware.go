package handlers

import (
	"net/http"

	"github.com/mohits-git/watch-expense/internal/ports"
	"github.com/mohits-git/watch-expense/internal/utils/authctx"
)

type AuthMiddleware struct {
	tokenProvider ports.TokenProvider
}

func NewAuthMiddleware(tokenProvider ports.TokenProvider) *AuthMiddleware {
	return &AuthMiddleware{tokenProvider: tokenProvider}
}

func (m *AuthMiddleware) Authenticated(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract the token from the Authorization header
		token := getBearerToken(r)
		if token == "" {
			writeError(w, http.StatusUnauthorized, "missing or invalid token")
			return
		}

		// validate the token
		userClaims, err := m.tokenProvider.ValidateToken(token)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "invalid token")
			return
		}

		ctx := authctx.WithUserClaims(r.Context(), &userClaims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *AuthMiddleware) WithToken(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := getBearerToken(r)
		if token == "" {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		_, err := m.tokenProvider.ValidateToken(token)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "invalid token")
			return
		}

		ctx := authctx.WithToken(r.Context(), token)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
