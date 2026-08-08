package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/ramekhchhoeng/car-rental/backend/internal/httputil"
	"github.com/ramekhchhoeng/car-rental/backend/internal/models"
)

type AdminHandler struct {
	DB *gorm.DB
}

// adminVehicleRow is a vehicle plus the count of bookings that would be
// disrupted if an admin took the listing down.
type adminVehicleRow struct {
	models.Vehicle
	ActiveBookings int64 `json:"active_bookings"`
}

// ListVehicles returns vehicles filtered by status (default: pending queue).
// GET /api/admin/vehicles?status=pending
func (h *AdminHandler) ListVehicles(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	if status == "" {
		status = string(models.VehiclePending)
	}

	var vehicles []models.Vehicle
	err := h.DB.Scopes(withReferences).Preload("Owner").
		Where("status = ?", status).
		Order("created_at ASC").
		Find(&vehicles).Error
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "could not load vehicles")
		return
	}

	httputil.JSON(w, http.StatusOK, h.withActiveBookings(vehicles))
}

// withActiveBookings attaches the live booking count to each vehicle using a
// single grouped query rather than one query per vehicle.
func (h *AdminHandler) withActiveBookings(vehicles []models.Vehicle) []adminVehicleRow {
	rows := make([]adminVehicleRow, len(vehicles))
	for i, v := range vehicles {
		rows[i] = adminVehicleRow{Vehicle: v}
	}
	if len(vehicles) == 0 {
		return rows
	}

	ids := make([]uuid.UUID, len(vehicles))
	for i, v := range vehicles {
		ids[i] = v.ID
	}

	var counts []struct {
		VehicleID uuid.UUID
		Total     int64
	}
	h.DB.Model(&models.Booking{}).
		Select("vehicle_id, COUNT(*) AS total").
		Where("vehicle_id IN ? AND status IN ?", ids,
			[]models.BookingStatus{models.BookingRequested, models.BookingConfirmed}).
		Group("vehicle_id").
		Scan(&counts)

	byVehicle := make(map[uuid.UUID]int64, len(counts))
	for _, c := range counts {
		byVehicle[c.VehicleID] = c.Total
	}
	for i := range rows {
		rows[i].ActiveBookings = byVehicle[rows[i].ID]
	}
	return rows
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
// It also serves as the take-down action for an already-approved listing.
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

// adminUserRow lists the columns an admin may see. Selecting fields explicitly
// (rather than SELECT *) guarantees the password hash can never leak.
type adminUserRow struct {
	ID           uuid.UUID   `json:"id"`
	Email        string      `json:"email"`
	FullName     string      `json:"full_name"`
	Phone        string      `json:"phone"`
	Role         models.Role `json:"role"`
	CreatedAt    time.Time   `json:"created_at"`
	VehicleCount int64       `json:"vehicle_count"`
	BookingCount int64       `json:"booking_count"`
}

// ListUsers returns every user with how many vehicles they list and how many
// bookings they have made. Supports ?role= and ?q= (name or email search).
// GET /api/admin/users?role=owner&q=sok
func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	q := h.DB.Model(&models.User{}).
		Select(`users.id, users.email, users.full_name, users.phone, users.role, users.created_at,
			(SELECT COUNT(*) FROM vehicles WHERE vehicles.owner_id = users.id) AS vehicle_count,
			(SELECT COUNT(*) FROM bookings WHERE bookings.renter_id = users.id) AS booking_count`)

	if role := r.URL.Query().Get("role"); role != "" {
		q = q.Where("users.role = ?", role)
	}
	if search := strings.TrimSpace(r.URL.Query().Get("q")); search != "" {
		like := "%" + search + "%"
		q = q.Where("users.full_name ILIKE ? OR users.email ILIKE ?", like, like)
	}

	rows := []adminUserRow{}
	if err := q.Order("users.created_at DESC").Scan(&rows).Error; err != nil {
		httputil.Error(w, http.StatusInternalServerError, "could not load users")
		return
	}
	httputil.JSON(w, http.StatusOK, rows)
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
