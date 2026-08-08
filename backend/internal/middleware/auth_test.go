package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/ramekhchhoeng/car-rental/backend/internal/models"
)

const testSecret = "test-secret"

func authedRequest(t *testing.T, token string) *http.Request {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	if token != "" {
		req.AddCookie(&http.Cookie{Name: CookieName, Value: token})
	}
	return req
}

// okHandler records that the request made it through the middleware.
func okHandler(called *bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*called = true
		w.WriteHeader(http.StatusOK)
	})
}

func TestRequireAuth_ValidToken(t *testing.T) {
	token, err := NewToken(testSecret, uuid.New(), models.RoleRenter)
	if err != nil {
		t.Fatalf("NewToken: %v", err)
	}

	called := false
	rec := httptest.NewRecorder()
	RequireAuth(testSecret)(okHandler(&called)).ServeHTTP(rec, authedRequest(t, token))

	if !called {
		t.Fatal("handler was not called with a valid token")
	}
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}

func TestRequireAuth_MissingCookie(t *testing.T) {
	called := false
	rec := httptest.NewRecorder()
	RequireAuth(testSecret)(okHandler(&called)).ServeHTTP(rec, authedRequest(t, ""))

	if called {
		t.Fatal("handler must not run without a token")
	}
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

func TestRequireAuth_ExpiredToken(t *testing.T) {
	claims := Claims{
		UserID: uuid.New().String(),
		Role:   string(models.RoleRenter),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(testSecret))
	if err != nil {
		t.Fatalf("sign: %v", err)
	}

	called := false
	rec := httptest.NewRecorder()
	RequireAuth(testSecret)(okHandler(&called)).ServeHTTP(rec, authedRequest(t, token))

	if called {
		t.Fatal("handler must not run with an expired token")
	}
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

func TestRequireAuth_WrongSecret(t *testing.T) {
	token, err := NewToken("attacker-secret", uuid.New(), models.RoleAdmin)
	if err != nil {
		t.Fatalf("NewToken: %v", err)
	}

	called := false
	rec := httptest.NewRecorder()
	RequireAuth(testSecret)(okHandler(&called)).ServeHTTP(rec, authedRequest(t, token))

	if called {
		t.Fatal("handler must not run with a forged token")
	}
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

func TestRequireRole(t *testing.T) {
	tests := []struct {
		name       string
		userRole   models.Role
		allowed    []models.Role
		wantStatus int
	}{
		{"owner allowed on owner route", models.RoleOwner, []models.Role{models.RoleOwner}, http.StatusOK},
		{"renter blocked on owner route", models.RoleRenter, []models.Role{models.RoleOwner}, http.StatusForbidden},
		{"admin blocked on owner route", models.RoleAdmin, []models.Role{models.RoleOwner}, http.StatusForbidden},
		{"admin allowed on admin route", models.RoleAdmin, []models.Role{models.RoleAdmin}, http.StatusOK},
		{"multiple roles accepted", models.RoleRenter, []models.Role{models.RoleRenter, models.RoleOwner}, http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := NewToken(testSecret, uuid.New(), tt.userRole)
			if err != nil {
				t.Fatalf("NewToken: %v", err)
			}

			called := false
			rec := httptest.NewRecorder()
			handler := RequireAuth(testSecret)(RequireRole(tt.allowed...)(okHandler(&called)))
			handler.ServeHTTP(rec, authedRequest(t, token))

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if (rec.Code == http.StatusOK) != called {
				t.Errorf("handler called = %v, inconsistent with status %d", called, rec.Code)
			}
		})
	}
}
