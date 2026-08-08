package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

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

// withReferences loads the vocabulary a listing is written in. Every path that
// returns a vehicle uses it, so no response can arrive holding a bare make_id
// the client has no way to render.
func withReferences(db *gorm.DB) *gorm.DB {
	return db.Preload("Photos").Preload("Make").Preload("Model").
		Preload("Province").Preload("Features")
}

// ListPublic returns approved vehicles with optional filters.
// GET /api/vehicles?type=car&province_id=…&make_id=…&min_seats=5&features=gps,helmet&max_price=100
func (h *VehicleHandler) ListPublic(w http.ResponseWriter, r *http.Request) {
	q := h.DB.Scopes(withReferences).Where("status = ?", models.VehicleApproved)

	if t := r.URL.Query().Get("type"); t != "" {
		q = q.Where("type = ?", t)
	}
	if id, err := uuid.Parse(r.URL.Query().Get("province_id")); err == nil {
		q = q.Where("province_id = ?", id)
	}
	if id, err := uuid.Parse(r.URL.Query().Get("make_id")); err == nil {
		q = q.Where("make_id = ?", id)
	}
	if seats := r.URL.Query().Get("min_seats"); seats != "" {
		if v, err := strconv.Atoi(seats); err == nil {
			q = q.Where("seats >= ?", v)
		}
	}
	// features=gps,air-conditioning — every one listed must be present, not any
	// of them: a renter ticking two boxes is stating two requirements.
	if raw := r.URL.Query().Get("features"); raw != "" {
		codes := strings.Split(raw, ",")
		q = q.Where(`(
			SELECT COUNT(DISTINCT f.code) FROM vehicle_features vf
			JOIN features f ON f.id = vf.feature_id
			WHERE vf.vehicle_id = vehicles.id AND f.code IN ?
		) = ?`, codes, len(codes))
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
	if err := h.DB.Scopes(withReferences).Preload("Owner").First(&vehicle, "id = ?", id).Error; err != nil {
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
	Type         string   `json:"type"`
	MakeID       string   `json:"make_id"`
	ModelID      string   `json:"model_id"`
	ProvinceID   string   `json:"province_id"`
	Year         int      `json:"year"`
	Transmission string   `json:"transmission"`
	Seats        int      `json:"seats"`
	PricePerDay  float64  `json:"price_per_day"`
	Description  string   `json:"description"`
	FeatureIDs   []string `json:"feature_ids"`
}

// references are the rows a request pointed at, once they have been proven to
// exist and to agree with each other.
type references struct {
	makeID     uuid.UUID
	modelID    uuid.UUID
	provinceID uuid.UUID
	features   []models.Feature
}

// validate covers what can be judged from the request alone. Anything needing
// the database is in resolve, so a malformed body never reaches it.
func (req *vehicleRequest) validate() string {
	if req.Type != string(models.VehicleCar) && req.Type != string(models.VehicleMotorbike) {
		return "type must be car or motorbike"
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
	if req.Seats < 1 || req.Seats > 64 {
		return "seats must be between 1 and 64"
	}
	return ""
}

// resolve turns the ids in a request into rows, and refuses the combinations
// that would otherwise produce a nonsense listing: a retired make, a model
// belonging to a different manufacturer, a car listed with a motorbike's model.
func (h *VehicleHandler) resolve(req *vehicleRequest) (*references, string) {
	provinceID, err := uuid.Parse(req.ProvinceID)
	if err != nil {
		return nil, "province is required"
	}
	var province models.Province
	if err := h.DB.First(&province, "id = ?", provinceID).Error; err != nil {
		return nil, "that province does not exist"
	}
	if !province.Active {
		return nil, "that province is no longer available"
	}

	makeID, err := uuid.Parse(req.MakeID)
	if err != nil {
		return nil, "make is required"
	}
	var mk models.VehicleMake
	if err := h.DB.First(&mk, "id = ?", makeID).Error; err != nil {
		return nil, "that make does not exist"
	}
	if !mk.Active {
		return nil, "that make is no longer available"
	}

	modelID, err := uuid.Parse(req.ModelID)
	if err != nil {
		return nil, "model is required"
	}
	var model models.VehicleModel
	if err := h.DB.First(&model, "id = ?", modelID).Error; err != nil {
		return nil, "that model does not exist"
	}
	if !model.Active {
		return nil, "that model is no longer available"
	}
	if model.MakeID != mk.ID {
		return nil, "that model belongs to a different make"
	}
	if string(model.Type) != req.Type {
		return nil, "that model is a " + string(model.Type) + ", not a " + req.Type
	}

	refs := &references{makeID: mk.ID, modelID: model.ID, provinceID: province.ID}

	if len(req.FeatureIDs) > 0 {
		ids := make([]uuid.UUID, 0, len(req.FeatureIDs))
		for _, raw := range req.FeatureIDs {
			id, err := uuid.Parse(raw)
			if err != nil {
				return nil, "invalid feature id"
			}
			ids = append(ids, id)
		}
		if err := h.DB.Where("id IN ? AND active", ids).Find(&refs.features).Error; err != nil {
			return nil, "could not load features"
		}
		if len(refs.features) != len(ids) {
			return nil, "one of those features does not exist"
		}
	}

	return refs, ""
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

	refs, msg := h.resolve(&req)
	if msg != "" {
		httputil.Error(w, http.StatusBadRequest, msg)
		return
	}

	vehicle := models.Vehicle{
		OwnerID:      middleware.UserIDFrom(r.Context()),
		Type:         models.VehicleType(req.Type),
		MakeID:       &refs.makeID,
		ModelID:      &refs.modelID,
		ProvinceID:   &refs.provinceID,
		Year:         req.Year,
		Transmission: models.Transmission(req.Transmission),
		Seats:        req.Seats,
		PricePerDay:  req.PricePerDay,
		Description:  req.Description,
		Features:     refs.features,
		Status:       models.VehiclePending,
	}
	if err := h.DB.Create(&vehicle).Error; err != nil {
		httputil.Error(w, http.StatusInternalServerError, "could not create vehicle")
		return
	}
	h.DB.Scopes(withReferences).First(&vehicle, "id = ?", vehicle.ID)
	httputil.JSON(w, http.StatusCreated, vehicle)
}

// ListOwn returns all vehicles belonging to the logged-in owner.
func (h *VehicleHandler) ListOwn(w http.ResponseWriter, r *http.Request) {
	var vehicles []models.Vehicle
	err := h.DB.Scopes(withReferences).
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
	if err := h.DB.Scopes(withReferences).First(&vehicle, "id = ?", id).Error; err != nil {
		httputil.Error(w, http.StatusNotFound, "vehicle not found")
		return nil, false
	}
	if vehicle.OwnerID != middleware.UserIDFrom(r.Context()) {
		httputil.Error(w, http.StatusForbidden, "not your vehicle")
		return nil, false
	}
	return &vehicle, true
}

// UpdateOwn edits an owner's vehicle. Editing a rejected listing resubmits it:
// a rejection names something to fix, so the act of fixing it is the resubmit,
// and the queue is where an admin decides whether it was fixed. Without this a
// rejection was permanent no matter what the owner changed.
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

	refs, msg := h.resolve(&req)
	if msg != "" {
		httputil.Error(w, http.StatusBadRequest, msg)
		return
	}

	vehicle.Type = models.VehicleType(req.Type)
	vehicle.MakeID = &refs.makeID
	vehicle.ModelID = &refs.modelID
	vehicle.ProvinceID = &refs.provinceID
	vehicle.Year = req.Year
	vehicle.Transmission = models.Transmission(req.Transmission)
	vehicle.Seats = req.Seats
	vehicle.PricePerDay = req.PricePerDay
	vehicle.Description = req.Description

	if vehicle.Status == models.VehicleRejected {
		vehicle.Status = models.VehiclePending
		// The reason described the old version. Keeping it would leave the queue
		// showing a complaint about photos that have since been replaced.
		vehicle.RejectionReason = ""
	}

	if err := h.DB.Save(vehicle).Error; err != nil {
		httputil.Error(w, http.StatusInternalServerError, "could not update vehicle")
		return
	}
	// Replace rather than append: unticking a feature has to remove it.
	if err := h.DB.Model(vehicle).Association("Features").Replace(refs.features); err != nil {
		httputil.Error(w, http.StatusInternalServerError, "could not update features")
		return
	}

	h.DB.Scopes(withReferences).First(vehicle, "id = ?", vehicle.ID)
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
