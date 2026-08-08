package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/ramekhchhoeng/car-rental/backend/internal/httputil"
	"github.com/ramekhchhoeng/car-rental/backend/internal/middleware"
	"github.com/ramekhchhoeng/car-rental/backend/internal/models"
)

type VehicleHandler struct {
	DB *gorm.DB
}

// ListPublic returns approved vehicles with optional filters.
// GET /api/vehicles?type=car&location=Phnom+Penh&min_price=10&max_price=100
func (h *VehicleHandler) ListPublic(w http.ResponseWriter, r *http.Request) {
	q := h.DB.Preload("Photos").Where("status = ?", models.VehicleApproved)

	if t := r.URL.Query().Get("type"); t != "" {
		q = q.Where("type = ?", t)
	}
	if loc := r.URL.Query().Get("location"); loc != "" {
		q = q.Where("location ILIKE ?", "%"+loc+"%")
	}
	if min := r.URL.Query().Get("min_price"); min != "" {
		if v, err := strconv.ParseFloat(min, 64); err == nil {
			q = q.Where("price_per_day >= ?", v)
		}
	}
	if max := r.URL.Query().Get("max_price"); max != "" {
		if v, err := strconv.ParseFloat(max, 64); err == nil {
			q = q.Where("price_per_day <= ?", v)
		}
	}

	var vehicles []models.Vehicle
	if err := q.Order("created_at DESC").Find(&vehicles).Error; err != nil {
		httputil.Error(w, http.StatusInternalServerError, "could not load vehicles")
		return
	}
	httputil.JSON(w, http.StatusOK, vehicles)
}

// GetPublic returns one vehicle. Approved is public; owners and admins can
// also see their own / any non-approved vehicle.
func (h *VehicleHandler) GetPublic(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid vehicle id")
		return
	}

	var vehicle models.Vehicle
	if err := h.DB.Preload("Photos").Preload("Owner").First(&vehicle, "id = ?", id).Error; err != nil {
		httputil.Error(w, http.StatusNotFound, "vehicle not found")
		return
	}

	if vehicle.Status != models.VehicleApproved {
		userID := middleware.UserIDFrom(r.Context())
		role := middleware.RoleFrom(r.Context())
		if role != models.RoleAdmin && userID != vehicle.OwnerID {
			httputil.Error(w, http.StatusNotFound, "vehicle not found")
			return
		}
	}

	// Average rating from completed bookings' reviews.
	var avg struct {
		Avg   float64
		Count int64
	}
	h.DB.Model(&models.Review{}).
		Select("COALESCE(AVG(rating),0) AS avg, COUNT(*) AS count").
		Joins("JOIN bookings ON bookings.id = reviews.booking_id").
		Where("bookings.vehicle_id = ?", vehicle.ID).
		Scan(&avg)

	httputil.JSON(w, http.StatusOK, map[string]any{
		"vehicle":      vehicle,
		"avg_rating":   avg.Avg,
		"review_count": avg.Count,
	})
}

type vehicleRequest struct {
	Type         string  `json:"type"`
	Make         string  `json:"make"`
	Model        string  `json:"model"`
	Year         int     `json:"year"`
	Transmission string  `json:"transmission"`
	Seats        int     `json:"seats"`
	PricePerDay  float64 `json:"price_per_day"`
	Location     string  `json:"location"`
	Description  string  `json:"description"`
}

func (req *vehicleRequest) validate() string {
	if req.Type != string(models.VehicleCar) && req.Type != string(models.VehicleMotorbike) {
		return "type must be car or motorbike"
	}
	if req.Make == "" || req.Model == "" {
		return "make and model are required"
	}
	if req.Year < 1980 || req.Year > 2100 {
		return "year looks invalid"
	}
	if req.Transmission != string(models.TransmissionAuto) && req.Transmission != string(models.TransmissionManual) {
		return "transmission must be auto or manual"
	}
	if req.PricePerDay <= 0 {
		return "price per day must be positive"
	}
	if req.Location == "" {
		return "location is required"
	}
	return ""
}

// CreateOwn creates a vehicle for the logged-in owner (status starts pending).
func (h *VehicleHandler) CreateOwn(w http.ResponseWriter, r *http.Request) {
	var req vehicleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if msg := req.validate(); msg != "" {
		httputil.Error(w, http.StatusBadRequest, msg)
		return
	}

	vehicle := models.Vehicle{
		OwnerID:      middleware.UserIDFrom(r.Context()),
		Type:         models.VehicleType(req.Type),
		Make:         req.Make,
		Model:        req.Model,
		Year:         req.Year,
		Transmission: models.Transmission(req.Transmission),
		Seats:        req.Seats,
		PricePerDay:  req.PricePerDay,
		Location:     req.Location,
		Description:  req.Description,
		Status:       models.VehiclePending,
	}
	if err := h.DB.Create(&vehicle).Error; err != nil {
		httputil.Error(w, http.StatusInternalServerError, "could not create vehicle")
		return
	}
	httputil.JSON(w, http.StatusCreated, vehicle)
}

// ListOwn returns all vehicles belonging to the logged-in owner.
func (h *VehicleHandler) ListOwn(w http.ResponseWriter, r *http.Request) {
	var vehicles []models.Vehicle
	err := h.DB.Preload("Photos").
		Where("owner_id = ?", middleware.UserIDFrom(r.Context())).
		Order("created_at DESC").Find(&vehicles).Error
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "could not load vehicles")
		return
	}
	httputil.JSON(w, http.StatusOK, vehicles)
}

// ownedVehicle loads a vehicle and checks it belongs to the logged-in owner.
func (h *VehicleHandler) ownedVehicle(w http.ResponseWriter, r *http.Request) (*models.Vehicle, bool) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid vehicle id")
		return nil, false
	}
	var vehicle models.Vehicle
	if err := h.DB.Preload("Photos").First(&vehicle, "id = ?", id).Error; err != nil {
		httputil.Error(w, http.StatusNotFound, "vehicle not found")
		return nil, false
	}
	if vehicle.OwnerID != middleware.UserIDFrom(r.Context()) {
		httputil.Error(w, http.StatusForbidden, "not your vehicle")
		return nil, false
	}
	return &vehicle, true
}

// UpdateOwn edits an owner's vehicle.
func (h *VehicleHandler) UpdateOwn(w http.ResponseWriter, r *http.Request) {
	vehicle, ok := h.ownedVehicle(w, r)
	if !ok {
		return
	}

	var req vehicleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if msg := req.validate(); msg != "" {
		httputil.Error(w, http.StatusBadRequest, msg)
		return
	}

	vehicle.Type = models.VehicleType(req.Type)
	vehicle.Make = req.Make
	vehicle.Model = req.Model
	vehicle.Year = req.Year
	vehicle.Transmission = models.Transmission(req.Transmission)
	vehicle.Seats = req.Seats
	vehicle.PricePerDay = req.PricePerDay
	vehicle.Location = req.Location
	vehicle.Description = req.Description

	if err := h.DB.Save(vehicle).Error; err != nil {
		httputil.Error(w, http.StatusInternalServerError, "could not update vehicle")
		return
	}
	httputil.JSON(w, http.StatusOK, vehicle)
}

// DeleteOwn removes an owner's vehicle.
func (h *VehicleHandler) DeleteOwn(w http.ResponseWriter, r *http.Request) {
	vehicle, ok := h.ownedVehicle(w, r)
	if !ok {
		return
	}
	if err := h.DB.Select("Photos").Delete(vehicle).Error; err != nil {
		httputil.Error(w, http.StatusInternalServerError, "could not delete vehicle")
		return
	}
	httputil.JSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
