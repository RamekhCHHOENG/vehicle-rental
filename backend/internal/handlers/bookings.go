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
	"github.com/ramekhchhoeng/car-rental/backend/internal/middleware"
	"github.com/ramekhchhoeng/car-rental/backend/internal/models"
)

type BookingHandler struct {
	DB *gorm.DB
}

const dateLayout = "2006-01-02"

type createBookingRequest struct {
	VehicleID string `json:"vehicle_id"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

// Create makes a booking request (renter only). Price is computed
// server-side so the client can never manipulate it.
func (h *BookingHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createBookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	vehicleID, err := uuid.Parse(req.VehicleID)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid vehicle id")
		return
	}
	start, err := time.Parse(dateLayout, req.StartDate)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "start_date must be YYYY-MM-DD")
		return
	}
	end, err := time.Parse(dateLayout, req.EndDate)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "end_date must be YYYY-MM-DD")
		return
	}

	today := time.Now().Truncate(24 * time.Hour)
	if start.Before(today) {
		httputil.Error(w, http.StatusBadRequest, "pick-up date cannot be in the past")
		return
	}
	days := RentalDays(start, end)
	if days < 1 {
		httputil.Error(w, http.StatusBadRequest, "return date must be after pick-up date")
		return
	}

	var vehicle models.Vehicle
	if err := h.DB.First(&vehicle, "id = ?", vehicleID).Error; err != nil {
		httputil.Error(w, http.StatusNotFound, "vehicle not found")
		return
	}
	if vehicle.Status != models.VehicleApproved {
		httputil.Error(w, http.StatusBadRequest, "this vehicle is not available for booking")
		return
	}

	// Refuse dates that clash with an already-confirmed booking.
	var confirmed []models.Booking
	if err := h.DB.Where("vehicle_id = ? AND status = ?", vehicle.ID, models.BookingConfirmed).Find(&confirmed).Error; err != nil {
		httputil.Error(w, http.StatusInternalServerError, "could not check availability")
		return
	}
	for _, b := range confirmed {
		if DatesOverlap(start, end, b.StartDate, b.EndDate) {
			httputil.Error(w, http.StatusConflict, "the vehicle is already booked for those dates")
			return
		}
	}

	booking := models.Booking{
		VehicleID:  vehicle.ID,
		RenterID:   middleware.UserIDFrom(r.Context()),
		StartDate:  start,
		EndDate:    end,
		TotalPrice: RentalTotal(vehicle.PricePerDay, days),
		Status:     models.BookingRequested,
	}
	if err := h.DB.Create(&booking).Error; err != nil {
		httputil.Error(w, http.StatusInternalServerError, "could not create booking")
		return
	}
	httputil.JSON(w, http.StatusCreated, booking)
}

// ListOwn returns the renter's bookings, newest first.
func (h *BookingHandler) ListOwn(w http.ResponseWriter, r *http.Request) {
	var bookings []models.Booking
	err := h.DB.Preload("Vehicle").Preload("Vehicle.Photos").
		Preload("Vehicle.Make").Preload("Vehicle.Model").Preload("Vehicle.Province").
		Preload("Review").
		Where("renter_id = ?", middleware.UserIDFrom(r.Context())).
		Order("created_at DESC").Find(&bookings).Error
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "could not load bookings")
		return
	}
	httputil.JSON(w, http.StatusOK, bookings)
}

// Cancel lets the renter cancel a requested or confirmed booking.
func (h *BookingHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	booking, ok := h.renterBooking(w, r)
	if !ok {
		return
	}
	if booking.Status != models.BookingRequested && booking.Status != models.BookingConfirmed {
		httputil.Error(w, http.StatusBadRequest, "only requested or confirmed bookings can be cancelled")
		return
	}
	booking.Status = models.BookingCancelled
	if err := h.DB.Save(booking).Error; err != nil {
		httputil.Error(w, http.StatusInternalServerError, "could not cancel booking")
		return
	}
	httputil.JSON(w, http.StatusOK, booking)
}

// Review lets the renter rate a completed booking once.
func (h *BookingHandler) Review(w http.ResponseWriter, r *http.Request) {
	booking, ok := h.renterBooking(w, r)
	if !ok {
		return
	}
	if booking.Status != models.BookingCompleted {
		httputil.Error(w, http.StatusBadRequest, "you can only review completed rentals")
		return
	}

	var req struct {
		Rating  int    `json:"rating"`
		Comment string `json:"comment"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.Rating < 1 || req.Rating > 5 {
		httputil.Error(w, http.StatusBadRequest, "rating must be between 1 and 5")
		return
	}

	review := models.Review{
		BookingID: booking.ID,
		Rating:    req.Rating,
		Comment:   strings.TrimSpace(req.Comment),
	}
	if err := h.DB.Create(&review).Error; err != nil {
		httputil.Error(w, http.StatusConflict, "this booking already has a review")
		return
	}
	httputil.JSON(w, http.StatusCreated, review)
}

// ListForOwner returns bookings on any of the owner's vehicles.
func (h *BookingHandler) ListForOwner(w http.ResponseWriter, r *http.Request) {
	var bookings []models.Booking
	err := h.DB.Preload("Vehicle").Preload("Vehicle.Make").Preload("Vehicle.Model").
		Preload("Vehicle.Province").Preload("Renter").
		Joins("JOIN vehicles ON vehicles.id = bookings.vehicle_id").
		Where("vehicles.owner_id = ?", middleware.UserIDFrom(r.Context())).
		Order("bookings.created_at DESC").Find(&bookings).Error
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "could not load bookings")
		return
	}
	httputil.JSON(w, http.StatusOK, bookings)
}

// Confirm lets the owner accept a requested booking,
// unless it overlaps a booking they already confirmed.
func (h *BookingHandler) Confirm(w http.ResponseWriter, r *http.Request) {
	booking, ok := h.ownerBooking(w, r)
	if !ok {
		return
	}
	if booking.Status != models.BookingRequested {
		httputil.Error(w, http.StatusBadRequest, "only requested bookings can be confirmed")
		return
	}

	var confirmed []models.Booking
	if err := h.DB.Where("vehicle_id = ? AND status = ?", booking.VehicleID, models.BookingConfirmed).Find(&confirmed).Error; err != nil {
		httputil.Error(w, http.StatusInternalServerError, "could not check availability")
		return
	}
	for _, other := range confirmed {
		if DatesOverlap(booking.StartDate, booking.EndDate, other.StartDate, other.EndDate) {
			httputil.Error(w, http.StatusConflict, "you already confirmed another booking for those dates")
			return
		}
	}

	booking.Status = models.BookingConfirmed
	if err := h.DB.Save(booking).Error; err != nil {
		httputil.Error(w, http.StatusInternalServerError, "could not confirm booking")
		return
	}
	httputil.JSON(w, http.StatusOK, booking)
}

// RejectBooking lets the owner decline a requested booking.
func (h *BookingHandler) RejectBooking(w http.ResponseWriter, r *http.Request) {
	booking, ok := h.ownerBooking(w, r)
	if !ok {
		return
	}
	if booking.Status != models.BookingRequested {
		httputil.Error(w, http.StatusBadRequest, "only requested bookings can be rejected")
		return
	}
	booking.Status = models.BookingRejected
	if err := h.DB.Save(booking).Error; err != nil {
		httputil.Error(w, http.StatusInternalServerError, "could not reject booking")
		return
	}
	httputil.JSON(w, http.StatusOK, booking)
}

// Complete lets the owner mark a confirmed booking as finished (vehicle returned).
func (h *BookingHandler) Complete(w http.ResponseWriter, r *http.Request) {
	booking, ok := h.ownerBooking(w, r)
	if !ok {
		return
	}
	if booking.Status != models.BookingConfirmed {
		httputil.Error(w, http.StatusBadRequest, "only confirmed bookings can be completed")
		return
	}
	booking.Status = models.BookingCompleted
	if err := h.DB.Save(booking).Error; err != nil {
		httputil.Error(w, http.StatusInternalServerError, "could not complete booking")
		return
	}
	httputil.JSON(w, http.StatusOK, booking)
}

// renterBooking loads a booking and checks it belongs to the logged-in renter.
func (h *BookingHandler) renterBooking(w http.ResponseWriter, r *http.Request) (*models.Booking, bool) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid booking id")
		return nil, false
	}
	var booking models.Booking
	if err := h.DB.First(&booking, "id = ?", id).Error; err != nil {
		httputil.Error(w, http.StatusNotFound, "booking not found")
		return nil, false
	}
	if booking.RenterID != middleware.UserIDFrom(r.Context()) {
		httputil.Error(w, http.StatusForbidden, "not your booking")
		return nil, false
	}
	return &booking, true
}

// ownerBooking loads a booking and checks the vehicle belongs to the logged-in owner.
func (h *BookingHandler) ownerBooking(w http.ResponseWriter, r *http.Request) (*models.Booking, bool) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid booking id")
		return nil, false
	}
	var booking models.Booking
	if err := h.DB.Preload("Vehicle").First(&booking, "id = ?", id).Error; err != nil {
		httputil.Error(w, http.StatusNotFound, "booking not found")
		return nil, false
	}
	if booking.Vehicle == nil || booking.Vehicle.OwnerID != middleware.UserIDFrom(r.Context()) {
		httputil.Error(w, http.StatusForbidden, "this booking is not on your vehicle")
		return nil, false
	}
	return &booking, true
}
