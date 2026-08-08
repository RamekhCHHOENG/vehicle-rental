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
	ID      uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	OwnerID uuid.UUID `gorm:"type:uuid;not null;index" json:"owner_id"`
	Owner   *User     `gorm:"foreignKey:OwnerID" json:"owner,omitempty"`

	Type VehicleType `gorm:"type:varchar(10);not null" json:"type"`

	// Make, model and province were free text until the reference tables in
	// metadata.go replaced them. They are pointers because the migration adds
	// the columns to rows that already exist: nullable here, required by
	// vehicleRequest.validate, so a legacy row that resisted backfill keeps the
	// service running instead of stopping it.
	MakeID     *uuid.UUID    `gorm:"type:uuid;index" json:"make_id"`
	Make       *VehicleMake  `gorm:"foreignKey:MakeID" json:"make,omitempty"`
	ModelID    *uuid.UUID    `gorm:"type:uuid;index" json:"model_id"`
	Model      *VehicleModel `gorm:"foreignKey:ModelID" json:"model,omitempty"`
	ProvinceID *uuid.UUID    `gorm:"type:uuid;index" json:"province_id"`
	Province   *Province     `gorm:"foreignKey:ProvinceID" json:"province,omitempty"`

	Year         int          `gorm:"not null" json:"year"`
	Transmission Transmission `gorm:"type:varchar(10);not null" json:"transmission"`
	Seats        int          `json:"seats"`
	PricePerDay  float64      `gorm:"type:numeric(10,2);not null" json:"price_per_day"`
	Description  string       `json:"description"`

	Features []Feature `gorm:"many2many:vehicle_features;constraint:OnDelete:CASCADE" json:"features"`

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
