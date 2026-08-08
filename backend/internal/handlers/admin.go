package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/ramekhchhoeng/car-rental/backend/internal/httputil"
	"github.com/ramekhchhoeng/car-rental/backend/internal/models"
)

type AdminHandler struct {
	DB *gorm.DB
}

// ListVehicles returns vehicles filtered by status (default: pending queue).
// GET /api/admin/vehicles?status=pending
func (h *AdminHandler) ListVehicles(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	if status == "" {
		status = string(models.VehiclePending)
	}

	var vehicles []models.Vehicle
	err := h.DB.Preload("Photos").Preload("Owner").
		Where("status = ?", status).
		Order("created_at ASC").
		Find(&vehicles).Error
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "could not load vehicles")
		return
	}
	httputil.JSON(w, http.StatusOK, vehicles)
}

func (h *AdminHandler) vehicleByID(w http.ResponseWriter, r *http.Request) (*models.Vehicle, bool) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid vehicle id")
		return nil, false
	}
	var vehicle models.Vehicle
	if err := h.DB.First(&vehicle, "id = ?", id).Error; err != nil {
		httputil.Error(w, http.StatusNotFound, "vehicle not found")
		return nil, false
	}
	return &vehicle, true
}

// Approve makes a vehicle publicly visible.
func (h *AdminHandler) Approve(w http.ResponseWriter, r *http.Request) {
	vehicle, ok := h.vehicleByID(w, r)
	if !ok {
		return
	}
	vehicle.Status = models.VehicleApproved
	vehicle.RejectionReason = ""
	if err := h.DB.Save(vehicle).Error; err != nil {
		httputil.Error(w, http.StatusInternalServerError, "could not approve vehicle")
		return
	}
	httputil.JSON(w, http.StatusOK, vehicle)
}

// Reject hides a vehicle and records why, so the owner can fix and relist.
func (h *AdminHandler) Reject(w http.ResponseWriter, r *http.Request) {
	vehicle, ok := h.vehicleByID(w, r)
	if !ok {
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.Reason) == "" {
		httputil.Error(w, http.StatusBadRequest, "a rejection reason is required")
		return
	}

	vehicle.Status = models.VehicleRejected
	vehicle.RejectionReason = strings.TrimSpace(req.Reason)
	if err := h.DB.Save(vehicle).Error; err != nil {
		httputil.Error(w, http.StatusInternalServerError, "could not reject vehicle")
		return
	}
	httputil.JSON(w, http.StatusOK, vehicle)
}

// ListUsers returns all users (newest first).
func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	var users []models.User
	if err := h.DB.Order("created_at DESC").Find(&users).Error; err != nil {
		httputil.Error(w, http.StatusInternalServerError, "could not load users")
		return
	}
	httputil.JSON(w, http.StatusOK, users)
}

// Stats returns the Key Metrics counters for the dashboard.
func (h *AdminHandler) Stats(w http.ResponseWriter, r *http.Request) {
	var stats struct {
		TotalUsers        int64 `json:"total_users"`
		TotalVehicles     int64 `json:"total_vehicles"`
		PendingVehicles   int64 `json:"pending_vehicles"`
		ApprovedVehicles  int64 `json:"approved_vehicles"`
		TotalBookings     int64 `json:"total_bookings"`
		CompletedBookings int64 `json:"completed_bookings"`
	}

	h.DB.Model(&models.User{}).Count(&stats.TotalUsers)
	h.DB.Model(&models.Vehicle{}).Count(&stats.TotalVehicles)
	h.DB.Model(&models.Vehicle{}).Where("status = ?", models.VehiclePending).Count(&stats.PendingVehicles)
	h.DB.Model(&models.Vehicle{}).Where("status = ?", models.VehicleApproved).Count(&stats.ApprovedVehicles)
	h.DB.Model(&models.Booking{}).Count(&stats.TotalBookings)
	h.DB.Model(&models.Booking{}).Where("status = ?", models.BookingCompleted).Count(&stats.CompletedBookings)

	httputil.JSON(w, http.StatusOK, stats)
}
