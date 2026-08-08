package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/ramekhchhoeng/car-rental/backend/internal/httputil"
	"github.com/ramekhchhoeng/car-rental/backend/internal/models"
)

type contextKey string

const (
	ctxUserID contextKey = "userID"
	ctxRole   contextKey = "role"
)

const CookieName = "auth_token"

type Claims struct {
	UserID string `json:"uid"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// NewToken creates a signed JWT for the given user, valid for 24h.
func NewToken(secret string, userID uuid.UUID, role models.Role) (string, error) {
	claims := Claims{
		UserID: userID.String(),
		Role:   string(role),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
}

// SetAuthCookie attaches the JWT as an httpOnly cookie.
func SetAuthCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   24 * 60 * 60,
	})
}

// ClearAuthCookie removes the auth cookie (logout).
func ClearAuthCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}

// RequireAuth validates the JWT cookie and puts user ID + role into the request context.
func RequireAuth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(CookieName)
			if err != nil {
				httputil.Error(w, http.StatusUnauthorized, "not logged in")
				return
			}

			claims := &Claims{}
			token, err := jwt.ParseWithClaims(cookie.Value, claims, func(t *jwt.Token) (any, error) {
				return []byte(secret), nil
			}, jwt.WithValidMethods([]string{"HS256"}))
			if err != nil || !token.Valid {
				httputil.Error(w, http.StatusUnauthorized, "invalid or expired session")
				return
			}

			userID, err := uuid.Parse(claims.UserID)
			if err != nil {
				httputil.Error(w, http.StatusUnauthorized, "invalid session")
				return
			}

			ctx := context.WithValue(r.Context(), ctxUserID, userID)
			ctx = context.WithValue(ctx, ctxRole, models.Role(claims.Role))
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OptionalAuth parses the JWT cookie when present but lets anonymous requests through.
func OptionalAuth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(CookieName)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			claims := &Claims{}
			token, err := jwt.ParseWithClaims(cookie.Value, claims, func(t *jwt.Token) (any, error) {
				return []byte(secret), nil
			}, jwt.WithValidMethods([]string{"HS256"}))
			if err != nil || !token.Valid {
				next.ServeHTTP(w, r)
				return
			}

			userID, err := uuid.Parse(claims.UserID)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			ctx := context.WithValue(r.Context(), ctxUserID, userID)
			ctx = context.WithValue(ctx, ctxRole, models.Role(claims.Role))
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole allows only the listed roles past. Must be nested inside RequireAuth.
func RequireRole(roles ...models.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role := RoleFrom(r.Context())
			for _, allowed := range roles {
				if role == allowed {
					next.ServeHTTP(w, r)
					return
				}
			}
			httputil.Error(w, http.StatusForbidden, "insufficient permissions")
		})
	}
}

// UserIDFrom returns the authenticated user's ID from the request context.
func UserIDFrom(ctx context.Context) uuid.UUID {
	id, _ := ctx.Value(ctxUserID).(uuid.UUID)
	return id
}

// RoleFrom returns the authenticated user's role from the request context.
func RoleFrom(ctx context.Context) models.Role {
	role, _ := ctx.Value(ctxRole).(models.Role)
	return role
}
