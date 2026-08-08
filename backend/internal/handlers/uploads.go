package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/ramekhchhoeng/car-rental/backend/internal/httputil"
	"github.com/ramekhchhoeng/car-rental/backend/internal/models"
)

const maxUploadSize = 10 << 20 // 10 MB

var allowedImageExts = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".webp": true,
}

// UploadPhoto handles POST /api/owner/vehicles/{id}/photos (multipart form, field "photo").
func (h *VehicleHandler) UploadPhoto(w http.ResponseWriter, r *http.Request) {
	vehicle, ok := h.ownedVehicle(w, r)
	if !ok {
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		httputil.Error(w, http.StatusBadRequest, "file too large (max 10 MB)")
		return
	}

	file, header, err := r.FormFile("photo")
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "photo file is required")
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !allowedImageExts[ext] {
		httputil.Error(w, http.StatusBadRequest, "only jpg, png, or webp images are allowed")
		return
	}

	if err := os.MkdirAll("uploads", 0o755); err != nil {
		httputil.Error(w, http.StatusInternalServerError, "could not store photo")
		return
	}

	filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	dst, err := os.Create(filepath.Join("uploads", filename))
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "could not store photo")
		return
	}
	defer dst.Close()
	if _, err := io.Copy(dst, file); err != nil {
		httputil.Error(w, http.StatusInternalServerError, "could not store photo")
		return
	}

	photo := models.VehiclePhoto{
		VehicleID: vehicle.ID,
		FilePath:  "/uploads/" + filename,
		SortOrder: len(vehicle.Photos),
	}
	if err := h.DB.Create(&photo).Error; err != nil {
		httputil.Error(w, http.StatusInternalServerError, "could not save photo record")
		return
	}
	httputil.JSON(w, http.StatusCreated, photo)
}

// DeletePhoto handles DELETE /api/owner/vehicles/{id}/photos/{photoId}.
func (h *VehicleHandler) DeletePhoto(w http.ResponseWriter, r *http.Request) {
	vehicle, ok := h.ownedVehicle(w, r)
	if !ok {
		return
	}

	photoID, err := uuid.Parse(chi.URLParam(r, "photoId"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid photo id")
		return
	}

	var photo models.VehiclePhoto
	if err := h.DB.First(&photo, "id = ? AND vehicle_id = ?", photoID, vehicle.ID).Error; err != nil {
		httputil.Error(w, http.StatusNotFound, "photo not found")
		return
	}

	os.Remove("." + photo.FilePath) // best-effort file cleanup
	if err := h.DB.Delete(&photo).Error; err != nil {
		httputil.Error(w, http.StatusInternalServerError, "could not delete photo")
		return
	}
	httputil.JSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
