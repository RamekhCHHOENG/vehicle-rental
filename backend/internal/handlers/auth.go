package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/ramekhchhoeng/car-rental/backend/internal/httputil"
	"github.com/ramekhchhoeng/car-rental/backend/internal/middleware"
	"github.com/ramekhchhoeng/car-rental/backend/internal/models"
)

type AuthHandler struct {
	DB        *gorm.DB
	JWTSecret string
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
	Phone    string `json:"phone"`
	Role     string `json:"role"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	if req.Email == "" || !strings.Contains(req.Email, "@") {
		httputil.Error(w, http.StatusBadRequest, "valid email is required")
		return
	}
	if len(req.Password) < 8 {
		httputil.Error(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}
	if strings.TrimSpace(req.FullName) == "" {
		httputil.Error(w, http.StatusBadRequest, "full name is required")
		return
	}
	role := models.Role(req.Role)
	if role != models.RoleRenter && role != models.RoleOwner {
		httputil.Error(w, http.StatusBadRequest, "role must be renter or owner")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "could not hash password")
		return
	}

	user := models.User{
		Email:        req.Email,
		PasswordHash: string(hash),
		FullName:     strings.TrimSpace(req.FullName),
		Phone:        strings.TrimSpace(req.Phone),
		Role:         role,
	}
	if err := h.DB.Create(&user).Error; err != nil {
		httputil.Error(w, http.StatusConflict, "email is already registered")
		return
	}

	h.issueSession(w, user)
	httputil.JSON(w, http.StatusCreated, user)
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	var user models.User
	err := h.DB.Where("email = ?", strings.ToLower(strings.TrimSpace(req.Email))).First(&user).Error
	if err != nil {
		httputil.Error(w, http.StatusUnauthorized, "wrong email or password")
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) != nil {
		httputil.Error(w, http.StatusUnauthorized, "wrong email or password")
		return
	}

	h.issueSession(w, user)
	httputil.JSON(w, http.StatusOK, user)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	middleware.ClearAuthCookie(w)
	httputil.JSON(w, http.StatusOK, map[string]string{"status": "logged out"})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFrom(r.Context())

	var user models.User
	if err := h.DB.First(&user, "id = ?", userID).Error; err != nil {
		httputil.Error(w, http.StatusUnauthorized, "user not found")
		return
	}
	httputil.JSON(w, http.StatusOK, user)
}

func (h *AuthHandler) issueSession(w http.ResponseWriter, user models.User) {
	token, err := middleware.NewToken(h.JWTSecret, user.ID, user.Role)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "could not create session")
		return
	}
	middleware.SetAuthCookie(w, token)
}
