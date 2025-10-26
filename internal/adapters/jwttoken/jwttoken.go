package jwttoken

import (
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/mohits-git/watch-expense/internal/domain"
	"github.com/mohits-git/watch-expense/internal/utils/apperr"
	"github.com/mohits-git/watch-expense/internal/utils/authctx"
)

type JWTService struct {
	secretKey string
	issuer    string
	audience  string
}

func NewJWTService(secretKey, issuer, audience string) *JWTService {
	return &JWTService{
		secretKey: secretKey,
		issuer:    issuer,
		audience:  audience,
	}
}

func (s *JWTService) GenerateToken(claims authctx.UserClaims) (string, error) {
	jwtClaims := jwt.MapClaims{
		"iss":     s.issuer,
		"aud":     s.audience,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
		"iat":     time.Now().Unix(),
		"user_id": claims.UserID,
		"sub":     claims.UserID,
		"role":    claims.Role,
		"name":    claims.Name,
		"email":   claims.Email,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims)
	return token.SignedString([]byte(s.secretKey))
}

func (s *JWTService) ValidateToken(tokenString string) (authctx.UserClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, apperr.NewAppError(apperr.ErrUnauthorized, "invalid token signature method", jwt.ErrTokenMalformed)
		}
		return []byte(s.secretKey), nil
	})
	if err != nil || !token.Valid {
		return authctx.UserClaims{}, apperr.NewAppError(apperr.ErrUnauthorized, "invalid token", jwt.ErrTokenMalformed)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return authctx.UserClaims{}, apperr.NewAppError(apperr.ErrUnauthorized, "invalid token claims", jwt.ErrTokenInvalidClaims)
	}
	if claims["iss"] != s.issuer || claims["aud"] != s.audience {
		return authctx.UserClaims{}, apperr.NewAppError(apperr.ErrUnauthorized, "invalid token issuer or audience", jwt.ErrTokenInvalidClaims)
	}

	userClaims, err := s.ExtractClaims(claims)
	return userClaims, err
}

func (s *JWTService) ExtractClaims(claims jwt.MapClaims) (authctx.UserClaims, error) {
	userID, ok := claims["user_id"].(string)
	if !ok {
		return authctx.UserClaims{}, apperr.NewAppError(apperr.ErrUnauthorized, "invalid user ID in token claims", jwt.ErrTokenInvalidClaims)
	}

	role, ok := claims["role"].(string)
	if !ok {
		return authctx.UserClaims{}, apperr.NewAppError(apperr.ErrUnauthorized, "invalid role in token claims", jwt.ErrTokenInvalidClaims)
	}

	return authctx.UserClaims{
		UserID: userID,
		Role:   domain.UserRole(role),
	}, nil
}
