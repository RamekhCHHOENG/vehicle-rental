package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BookingStatus string

const (
	BookingRequested BookingStatus = "requested"
	BookingConfirmed BookingStatus = "confirmed"
	BookingRejected  BookingStatus = "rejected"
	BookingCancelled BookingStatus = "cancelled"
	BookingCompleted BookingStatus = "completed"
)

type Booking struct {
	ID         uuid.UUID     `gorm:"type:uuid;primaryKey" json:"id"`
	VehicleID  uuid.UUID     `gorm:"type:uuid;not null;index" json:"vehicle_id"`
	Vehicle    *Vehicle      `gorm:"foreignKey:VehicleID" json:"vehicle,omitempty"`
	RenterID   uuid.UUID     `gorm:"type:uuid;not null;index" json:"renter_id"`
	Renter     *User         `gorm:"foreignKey:RenterID" json:"renter,omitempty"`
	StartDate  time.Time     `gorm:"type:date;not null" json:"start_date"`
	EndDate    time.Time     `gorm:"type:date;not null" json:"end_date"`
	TotalPrice float64       `gorm:"type:numeric(10,2);not null" json:"total_price"`
	Status     BookingStatus `gorm:"type:varchar(10);not null;default:requested;index" json:"status"`
	Review     *Review       `gorm:"foreignKey:BookingID" json:"review,omitempty"`
	CreatedAt  time.Time     `json:"created_at"`
	UpdatedAt  time.Time     `json:"updated_at"`
}

func (b *Booking) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}

type Review struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	BookingID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex" json:"booking_id"`
	Rating    int       `gorm:"not null;check:rating >= 1 AND rating <= 5" json:"rating"`
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"created_at"`
}

func (r *Review) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}
