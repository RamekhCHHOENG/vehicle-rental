package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type VehicleType string

const (
	VehicleCar       VehicleType = "car"
	VehicleMotorbike VehicleType = "motorbike"
)

type VehicleStatus string

const (
	VehiclePending  VehicleStatus = "pending"
	VehicleApproved VehicleStatus = "approved"
	VehicleRejected VehicleStatus = "rejected"
)

type Transmission string

const (
	TransmissionAuto   Transmission = "auto"
	TransmissionManual Transmission = "manual"
)

type Vehicle struct {
	ID              uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	OwnerID         uuid.UUID      `gorm:"type:uuid;not null;index" json:"owner_id"`
	Owner           *User          `gorm:"foreignKey:OwnerID" json:"owner,omitempty"`
	Type            VehicleType    `gorm:"type:varchar(10);not null" json:"type"`
	Make            string         `gorm:"not null" json:"make"`
	Model           string         `gorm:"not null" json:"model"`
	Year            int            `gorm:"not null" json:"year"`
	Transmission    Transmission   `gorm:"type:varchar(10);not null" json:"transmission"`
	Seats           int            `json:"seats"`
	PricePerDay     float64        `gorm:"type:numeric(10,2);not null" json:"price_per_day"`
	Location        string         `gorm:"not null" json:"location"`
	Description     string         `json:"description"`
	Status          VehicleStatus  `gorm:"type:varchar(10);not null;default:pending;index" json:"status"`
	RejectionReason string         `json:"rejection_reason,omitempty"`
	Photos          []VehiclePhoto `gorm:"foreignKey:VehicleID;constraint:OnDelete:CASCADE" json:"photos"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
}

func (v *Vehicle) BeforeCreate(tx *gorm.DB) error {
	if v.ID == uuid.Nil {
		v.ID = uuid.New()
	}
	return nil
}

type VehiclePhoto struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	VehicleID uuid.UUID `gorm:"type:uuid;not null;index" json:"vehicle_id"`
	FilePath  string    `gorm:"not null" json:"file_path"`
	SortOrder int       `json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
}

func (p *VehiclePhoto) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}
