package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/ramekhchhoeng/car-rental/backend/internal/httputil"
	"github.com/ramekhchhoeng/car-rental/backend/internal/models"
)

type MetadataHandler struct {
	DB *gorm.DB
}

type option struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

// vehicleTypes and transmissions are not in the database on purpose. Handlers
// branch on these values and the booking rules read them, so an admin renaming
// "car" at runtime would break code rather than data. They are served here
// anyway so the frontend has one source for every list a form needs, instead of
// a hardcoded copy that drifts from the Go constants.
var (
	vehicleTypes = []option{
		{string(models.VehicleCar), "Car"},
		{string(models.VehicleMotorbike), "Motorbike"},
	}
	transmissions = []option{
		{string(models.TransmissionAuto), "Automatic"},
		{string(models.TransmissionManual), "Manual"},
	}
	// Seat counts that occur in practice: two for a motorbike, up to a minibus.
	seatOptions = []int{2, 4, 5, 7, 8, 12, 16}
)

// makeRow is a make plus whether it has anything of each type, so a form can
// offer only the makes that sell what is being listed without downloading every
// model to work it out.
type makeRow struct {
	models.VehicleMake
	HasCars       bool `json:"has_cars"`
	HasMotorbikes bool `json:"has_motorbikes"`
}

// listMakes derives the type flags in the query rather than storing them, so
// they cannot drift from the models they describe.
func (h *MetadataHandler) listMakes() ([]makeRow, error) {
	var rows []makeRow
	err := h.DB.Model(&models.VehicleMake{}).
		Select(`vehicle_makes.*,
			EXISTS (SELECT 1 FROM vehicle_models m
				WHERE m.make_id = vehicle_makes.id AND m.active AND m.type = ?) AS has_cars,
			EXISTS (SELECT 1 FROM vehicle_models m
				WHERE m.make_id = vehicle_makes.id AND m.active AND m.type = ?) AS has_motorbikes`,
			models.VehicleCar, models.VehicleMotorbike).
		Where("vehicle_makes.active").
		Order("vehicle_makes.sort_order, vehicle_makes.name").
		Scan(&rows).Error
	return rows, err
}

// ListModels returns one make's models. Separate from /api/metadata because
// models are the part that scales: a few thousand makes is a small payload,
// their combined models are not.
// GET /api/makes/{id}/models?type=car
func (h *MetadataHandler) ListModels(w http.ResponseWriter, r *http.Request) {
	makeID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid make id")
		return
	}

	q := h.DB.Where("make_id = ? AND active", makeID)
	if t := r.URL.Query().Get("type"); t != "" {
		q = q.Where("type = ?", t)
	}

	var found []models.VehicleModel
	if err := q.Order("name").Find(&found).Error; err != nil {
		httputil.Error(w, http.StatusInternalServerError, "could not load models")
		return
	}
	httputil.JSON(w, http.StatusOK, found)
}

// Get returns every list the listing form and the browse filters need, in one
// request — so no page can be showing a vocabulary that another page has
// already moved on from. Makes come without their models; fetch those from
// ListModels once a make is chosen.
// GET /api/metadata
func (h *MetadataHandler) Get(w http.ResponseWriter, r *http.Request) {
	var provinces []models.Province
	if err := h.DB.Where("active").Order("sort_order, name_en").Find(&provinces).Error; err != nil {
		httputil.Error(w, http.StatusInternalServerError, "could not load provinces")
		return
	}

	makes, err := h.listMakes()
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "could not load makes")
		return
	}

	var features []models.Feature
	if err := h.DB.Where("active").Order("sort_order, name").Find(&features).Error; err != nil {
		httputil.Error(w, http.StatusInternalServerError, "could not load features")
		return
	}

	httputil.JSON(w, http.StatusOK, map[string]any{
		"provinces":     provinces,
		"makes":         makes,
		"features":      features,
		"vehicle_types": vehicleTypes,
		"transmissions": transmissions,
		"seat_options":  seatOptions,
	})
}
