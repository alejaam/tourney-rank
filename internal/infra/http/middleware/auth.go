package middleware

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/alejaam/tourney-rank/internal/domain/user"
	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const (
	UserContextKey contextKey = "user"
)

// UserInfo contains authenticated user information.
type UserInfo struct {
	ID   string
	Role user.Role
}

// Auth validates JWT tokens and adds user info to context.
func Auth(jwtSecret string, logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				logger.Debug("missing authorization header")
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				logger.Debug("invalid authorization header format")
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			tokenString := parts[1]

			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, errors.New("invalid signing method")
				}
				return []byte(jwtSecret), nil
			})

			if err != nil || !token.Valid {
				logger.Debug("invalid token", "error", err)
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				logger.Debug("invalid claims")
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			userID, ok := claims["sub"].(string)
			if !ok {
				logger.Debug("missing user id in claims")
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			roleStr, ok := claims["role"].(string)
			if !ok {
				logger.Debug("missing role in claims")
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			userInfo := &UserInfo{
				ID:   userID,
				Role: user.Role(roleStr),
			}

			ctx := context.WithValue(r.Context(), UserContextKey, userInfo)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// AdminOnly ensures the user has admin role.
func AdminOnly(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userInfo, ok := r.Context().Value(UserContextKey).(*UserInfo)
			if !ok {
				logger.Debug("user info not found in context")
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			if userInfo.Role != user.RoleAdmin {
				logger.Debug("user is not admin", "user_id", userInfo.ID, "role", userInfo.Role)
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetUserInfo retrieves user info from context.
func GetUserInfo(ctx context.Context) (*UserInfo, bool) {
	userInfo, ok := ctx.Value(UserContextKey).(*UserInfo)
	return userInfo, ok
}
