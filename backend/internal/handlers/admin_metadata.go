package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/ramekhchhoeng/car-rental/backend/internal/httputil"
	"github.com/ramekhchhoeng/car-rental/backend/internal/models"
)

// Admin management of the listing vocabulary. Two rules run through all of it:
//
//   - Nothing a listing points at can be deleted. Deleting a make would either
//     orphan the vehicles listed under it or silently rewrite what their owners
//     said. Deactivating hides it from every form while leaving those listings
//     truthful, so that is what a delete attempt is redirected to.
//   - Lists are returned in full, inactive rows included. The public
//     /api/metadata endpoint filters them out; the admin needs to see what it
//     has retired in order to bring it back.
type AdminMetadataHandler struct {
	DB *gorm.DB
}

// ListAll returns the small lists in one response, retired entries included.
//
// Makes are deliberately not here. After an NHTSA import there are a few
// thousand of them and tens of thousands of models between them, so they are
// paged and searched through ListMakes instead.
// GET /api/admin/metadata
func (h *AdminMetadataHandler) ListAll(w http.ResponseWriter, r *http.Request) {
	var provinces []models.Province
	var features []models.Feature

	h.DB.Order("sort_order, name_en").Find(&provinces)
	h.DB.Order("sort_order, name").Find(&features)

	// Usage counts turn "can I delete this?" into something visible rather than
	// something you discover by being refused.
	type usage struct {
		ID    uuid.UUID
		Count int64
	}
	counts := map[string]map[uuid.UUID]int64{}
	for _, spec := range []struct{ key, column string }{
		{"provinces", "province_id"},
	} {
		var rows []usage
		h.DB.Model(&models.Vehicle{}).
			Select(spec.column + " AS id, COUNT(*) AS count").
			Where(spec.column + " IS NOT NULL").
			Group(spec.column).Scan(&rows)
		m := map[uuid.UUID]int64{}
		for _, row := range rows {
			m[row.ID] = row.Count
		}
		counts[spec.key] = m
	}
	var featureRows []usage
	h.DB.Table("vehicle_features").
		Select("feature_id AS id, COUNT(*) AS count").Group("feature_id").Scan(&featureRows)
	featureCounts := map[uuid.UUID]int64{}
	for _, row := range featureRows {
		featureCounts[row.ID] = row.Count
	}
	counts["features"] = featureCounts

	httputil.JSON(w, http.StatusOK, map[string]any{
		"provinces": provinces,
		"features":  features,
		"usage":     counts,
	})
}

// adminMakeRow carries the two numbers that decide what an admin can do with a
// make: how many models hang off it, and how many listings depend on it.
type adminMakeRow struct {
	models.VehicleMake
	ModelCount   int64 `json:"model_count"`
	ListingCount int64 `json:"listing_count"`
}

// ListMakes pages and searches the makes. A thousand-row table needs a search
// box more than it needs to be complete.
// GET /api/admin/makes?q=toy&limit=50&offset=0
func (h *AdminMetadataHandler) ListMakes(w http.ResponseWriter, r *http.Request) {
	limit := intParam(r, "limit", 50, 200)
	offset := intParam(r, "offset", 0, 1_000_000)

	q := h.DB.Model(&models.VehicleMake{})
	if search := strings.TrimSpace(r.URL.Query().Get("q")); search != "" {
		q = q.Where("name ILIKE ?", "%"+search+"%")
	}

	var total int64
	q.Count(&total)

	var rows []adminMakeRow
	err := q.Select(`vehicle_makes.*,
			(SELECT COUNT(*) FROM vehicle_models m WHERE m.make_id = vehicle_makes.id) AS model_count,
			(SELECT COUNT(*) FROM vehicles v WHERE v.make_id = vehicle_makes.id) AS listing_count`).
		Order("vehicle_makes.sort_order, vehicle_makes.name").
		Limit(limit).Offset(offset).
		Scan(&rows).Error
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "could not load makes")
		return
	}

	httputil.JSON(w, http.StatusOK, map[string]any{
		"makes":  rows,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// adminModelRow is a model plus how many listings use it, so the same
// delete-or-retire decision can be made one level down.
type adminModelRow struct {
	models.VehicleModel
	ListingCount int64 `json:"listing_count"`
}

// ListModelsForMake returns one make's models, retired ones included — the
// public endpoint hides those, and the admin is the one who has to bring them
// back.
// GET /api/admin/makes/{id}/models
func (h *AdminMetadataHandler) ListModelsForMake(w http.ResponseWriter, r *http.Request) {
	makeID, ok := parseID(w, r)
	if !ok {
		return
	}

	var rows []adminModelRow
	err := h.DB.Model(&models.VehicleModel{}).
		Select(`vehicle_models.*,
			(SELECT COUNT(*) FROM vehicles v WHERE v.model_id = vehicle_models.id) AS listing_count`).
		Where("make_id = ?", makeID).
		Order("vehicle_models.type, vehicle_models.name").
		Scan(&rows).Error
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "could not load models")
		return
	}
	httputil.JSON(w, http.StatusOK, rows)
}

func intParam(r *http.Request, name string, fallback, max int) int {
	raw := r.URL.Query().Get(name)
	if raw == "" {
		return fallback
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v < 0 {
		return fallback
	}
	if v > max {
		return max
	}
	return v
}

// --- provinces ---

type provinceRequest struct {
	Code      string `json:"code"`
	NameEn    string `json:"name_en"`
	NameKm    string `json:"name_km"`
	SortOrder int    `json:"sort_order"`
	Active    *bool  `json:"active"`
}

func (h *AdminMetadataHandler) CreateProvince(w http.ResponseWriter, r *http.Request) {
	var req provinceRequest
	if !decode(w, r, &req) {
		return
	}
	if req.NameEn == "" {
		httputil.Error(w, http.StatusBadRequest, "name_en is required")
		return
	}
	if req.Code == "" {
		req.Code = slugify(req.NameEn)
	}

	province := models.Province{
		Code: req.Code, NameEn: req.NameEn, NameKm: req.NameKm,
		SortOrder: req.SortOrder, Active: boolOr(req.Active, true),
	}
	if err := h.DB.Create(&province).Error; err != nil {
		httputil.Error(w, http.StatusConflict, "a province with that code already exists")
		return
	}
	httputil.JSON(w, http.StatusCreated, province)
}

func (h *AdminMetadataHandler) UpdateProvince(w http.ResponseWriter, r *http.Request) {
	var province models.Province
	if !h.load(w, r, &province) {
		return
	}
	var req provinceRequest
	if !decode(w, r, &req) {
		return
	}
	if req.NameEn != "" {
		province.NameEn = req.NameEn
	}
	province.NameKm = req.NameKm
	province.SortOrder = req.SortOrder
	province.Active = boolOr(req.Active, province.Active)

	if err := h.DB.Save(&province).Error; err != nil {
		httputil.Error(w, http.StatusInternalServerError, "could not update province")
		return
	}
	httputil.JSON(w, http.StatusOK, province)
}

func (h *AdminMetadataHandler) DeleteProvince(w http.ResponseWriter, r *http.Request) {
	h.deleteIfUnused(w, r, &models.Province{}, "province_id", "province")
}

// --- makes and models ---

type makeRequest struct {
	Name      string `json:"name"`
	SortOrder int    `json:"sort_order"`
	Active    *bool  `json:"active"`
}

func (h *AdminMetadataHandler) CreateMake(w http.ResponseWriter, r *http.Request) {
	var req makeRequest
	if !decode(w, r, &req) {
		return
	}
	if req.Name == "" {
		httputil.Error(w, http.StatusBadRequest, "name is required")
		return
	}
	mk := models.VehicleMake{Name: req.Name, SortOrder: req.SortOrder, Active: boolOr(req.Active, true)}
	if err := h.DB.Create(&mk).Error; err != nil {
		httputil.Error(w, http.StatusConflict, "a make with that name already exists")
		return
	}
	httputil.JSON(w, http.StatusCreated, mk)
}

func (h *AdminMetadataHandler) UpdateMake(w http.ResponseWriter, r *http.Request) {
	var mk models.VehicleMake
	if !h.load(w, r, &mk) {
		return
	}
	var req makeRequest
	if !decode(w, r, &req) {
		return
	}
	if req.Name != "" {
		mk.Name = req.Name
	}
	mk.SortOrder = req.SortOrder
	mk.Active = boolOr(req.Active, mk.Active)

	if err := h.DB.Save(&mk).Error; err != nil {
		httputil.Error(w, http.StatusConflict, "could not update make")
		return
	}
	httputil.JSON(w, http.StatusOK, mk)
}

func (h *AdminMetadataHandler) DeleteMake(w http.ResponseWriter, r *http.Request) {
	h.deleteIfUnused(w, r, &models.VehicleMake{}, "make_id", "make")
}

type modelRequest struct {
	MakeID string `json:"make_id"`
	Name   string `json:"name"`
	Type   string `json:"type"`
	Active *bool  `json:"active"`
}

func (h *AdminMetadataHandler) CreateModel(w http.ResponseWriter, r *http.Request) {
	var req modelRequest
	if !decode(w, r, &req) {
		return
	}
	makeID, err := uuid.Parse(req.MakeID)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "make_id is required")
		return
	}
	if req.Name == "" {
		httputil.Error(w, http.StatusBadRequest, "name is required")
		return
	}
	if req.Type != string(models.VehicleCar) && req.Type != string(models.VehicleMotorbike) {
		httputil.Error(w, http.StatusBadRequest, "type must be car or motorbike")
		return
	}
	if err := h.DB.First(&models.VehicleMake{}, "id = ?", makeID).Error; err != nil {
		httputil.Error(w, http.StatusBadRequest, "that make does not exist")
		return
	}

	m := models.VehicleModel{
		MakeID: makeID, Name: req.Name,
		Type: models.VehicleType(req.Type), Active: boolOr(req.Active, true),
	}
	if err := h.DB.Create(&m).Error; err != nil {
		httputil.Error(w, http.StatusConflict, "that make already has a model with this name")
		return
	}
	httputil.JSON(w, http.StatusCreated, m)
}

func (h *AdminMetadataHandler) UpdateModel(w http.ResponseWriter, r *http.Request) {
	var m models.VehicleModel
	if !h.load(w, r, &m) {
		return
	}
	var req modelRequest
	if !decode(w, r, &req) {
		return
	}
	if req.Name != "" {
		m.Name = req.Name
	}
	if req.Type == string(models.VehicleCar) || req.Type == string(models.VehicleMotorbike) {
		m.Type = models.VehicleType(req.Type)
	}
	m.Active = boolOr(req.Active, m.Active)

	if err := h.DB.Save(&m).Error; err != nil {
		httputil.Error(w, http.StatusConflict, "could not update model")
		return
	}
	httputil.JSON(w, http.StatusOK, m)
}

func (h *AdminMetadataHandler) DeleteModel(w http.ResponseWriter, r *http.Request) {
	h.deleteIfUnused(w, r, &models.VehicleModel{}, "model_id", "model")
}

// --- features ---

type featureRequest struct {
	Code      string `json:"code"`
	Name      string `json:"name"`
	Icon      string `json:"icon"`
	AppliesTo string `json:"applies_to"`
	SortOrder int    `json:"sort_order"`
	Active    *bool  `json:"active"`
}

func (h *AdminMetadataHandler) CreateFeature(w http.ResponseWriter, r *http.Request) {
	var req featureRequest
	if !decode(w, r, &req) {
		return
	}
	if req.Name == "" {
		httputil.Error(w, http.StatusBadRequest, "name is required")
		return
	}
	if req.Code == "" {
		req.Code = slugify(req.Name)
	}
	f := models.Feature{
		Code: req.Code, Name: req.Name, Icon: req.Icon,
		AppliesTo: models.VehicleType(req.AppliesTo),
		SortOrder: req.SortOrder, Active: boolOr(req.Active, true),
	}
	if err := h.DB.Create(&f).Error; err != nil {
		httputil.Error(w, http.StatusConflict, "a feature with that code already exists")
		return
	}
	httputil.JSON(w, http.StatusCreated, f)
}

func (h *AdminMetadataHandler) UpdateFeature(w http.ResponseWriter, r *http.Request) {
	var f models.Feature
	if !h.load(w, r, &f) {
		return
	}
	var req featureRequest
	if !decode(w, r, &req) {
		return
	}
	if req.Name != "" {
		f.Name = req.Name
	}
	f.Icon = req.Icon
	f.AppliesTo = models.VehicleType(req.AppliesTo)
	f.SortOrder = req.SortOrder
	f.Active = boolOr(req.Active, f.Active)

	if err := h.DB.Save(&f).Error; err != nil {
		httputil.Error(w, http.StatusInternalServerError, "could not update feature")
		return
	}
	httputil.JSON(w, http.StatusOK, f)
}

func (h *AdminMetadataHandler) DeleteFeature(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	var count int64
	h.DB.Table("vehicle_features").Where("feature_id = ?", id).Count(&count)
	if count > 0 {
		httputil.Error(w, http.StatusConflict, inUseMessage("feature", count))
		return
	}
	if err := h.DB.Delete(&models.Feature{}, "id = ?", id).Error; err != nil {
		httputil.Error(w, http.StatusInternalServerError, "could not delete feature")
		return
	}
	httputil.JSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// --- shared helpers ---

func (h *AdminMetadataHandler) load(w http.ResponseWriter, r *http.Request, dest any) bool {
	id, ok := parseID(w, r)
	if !ok {
		return false
	}
	if err := h.DB.First(dest, "id = ?", id).Error; err != nil {
		httputil.Error(w, http.StatusNotFound, "not found")
		return false
	}
	return true
}

// deleteIfUnused removes the row only when no vehicle references it. The
// alternative — cascading, or nulling the column — would quietly change what a
// listing claims about itself, so a refusal plus a nudge to deactivate is the
// safer answer.
func (h *AdminMetadataHandler) deleteIfUnused(w http.ResponseWriter, r *http.Request, dest any, column, label string) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	var count int64
	h.DB.Model(&models.Vehicle{}).Where(column+" = ?", id).Count(&count)
	if count > 0 {
		httputil.Error(w, http.StatusConflict, inUseMessage(label, count))
		return
	}
	if err := h.DB.Delete(dest, "id = ?", id).Error; err != nil {
		httputil.Error(w, http.StatusInternalServerError, "could not delete "+label)
		return
	}
	httputil.JSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func inUseMessage(label string, count int64) string {
	plural := "s"
	if count == 1 {
		plural = ""
	}
	return "this " + label + " is used by " + strconv.FormatInt(count, 10) + " listing" + plural +
		". Deactivate it instead — it disappears from the forms and those listings stay as their owners wrote them."
}

func parseID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid id")
		return uuid.Nil, false
	}
	return id, true
}

// decodeOptional tolerates an empty body, for endpoints where sending nothing
// means "use the defaults" rather than "malformed request".
func decodeOptional(r *http.Request, dest any) error {
	err := json.NewDecoder(r.Body).Decode(dest)
	if err == io.EOF {
		return nil
	}
	return err
}

func decode(w http.ResponseWriter, r *http.Request, dest any) bool {
	if err := json.NewDecoder(r.Body).Decode(dest); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid JSON body")
		return false
	}
	return true
}

// boolOr keeps an omitted "active" from silently retiring a row: only an
// explicit false in the body deactivates.
func boolOr(v *bool, fallback bool) bool {
	if v == nil {
		return fallback
	}
	return *v
}

func slugify(s string) string {
	var b strings.Builder
	prevDash := false
	for _, r := range strings.ToLower(strings.TrimSpace(s)) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			prevDash = false
		case !prevDash && b.Len() > 0:
			b.WriteRune('-')
			prevDash = true
		}
	}
	return strings.TrimSuffix(b.String(), "-")
}
