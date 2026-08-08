package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// The tables in this file are the vocabulary the listings are written in:
// where a vehicle is, what it is, and what comes with it. They exist because
// free text does not survive contact with users — two owners typing the same
// province three different ways makes both listings unfindable.
//
// Every one of them is deactivated rather than deleted once a listing points
// at it. Retiring a make must not rewrite the history of vehicles that were
// honestly listed under it.

// Province is one of Cambodia's 25 provinces. Seeded, not user-created: the
// list is set by the government, not by us.
type Province struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Code      string    `gorm:"uniqueIndex;not null" json:"code"`
	NameEn    string    `gorm:"not null" json:"name_en"`
	NameKm    string    `gorm:"not null" json:"name_km"`
	SortOrder int       `json:"sort_order"`
	Active    bool      `gorm:"not null;default:true" json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (p *Province) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

// VehicleMake is a manufacturer. It is deliberately not typed car-or-motorbike:
// Honda and Suzuki sell both, so the type belongs on the model instead.
type VehicleMake struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Name      string         `gorm:"uniqueIndex;not null" json:"name"`
	SortOrder int            `json:"sort_order"`
	Active    bool           `gorm:"not null;default:true" json:"active"`
	Models    []VehicleModel `gorm:"foreignKey:MakeID;constraint:OnDelete:CASCADE" json:"models,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

func (m *VehicleMake) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}

// VehicleModel is one model from one make. Type is what lets the form show
// Honda's cars when the listing is a car and Honda's motorbikes when it is not.
type VehicleModel struct {
	ID        uuid.UUID   `gorm:"type:uuid;primaryKey" json:"id"`
	MakeID    uuid.UUID   `gorm:"type:uuid;not null;uniqueIndex:idx_model_make_name" json:"make_id"`
	Name      string      `gorm:"not null;uniqueIndex:idx_model_make_name" json:"name"`
	Type      VehicleType `gorm:"type:varchar(10);not null;index" json:"type"`
	Active    bool        `gorm:"not null;default:true" json:"active"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

func (m *VehicleModel) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}

// Feature is something a listing either includes or does not — air
// conditioning, a helmet, delivery. Renters filter on these, so they have to
// be a fixed vocabulary rather than a sentence in the description.
type Feature struct {
	ID        uuid.UUID   `gorm:"type:uuid;primaryKey" json:"id"`
	Code      string      `gorm:"uniqueIndex;not null" json:"code"`
	Name      string      `gorm:"not null" json:"name"`
	Icon      string      `json:"icon"`
	AppliesTo VehicleType `gorm:"type:varchar(10)" json:"applies_to,omitempty"`
	SortOrder int         `json:"sort_order"`
	Active    bool        `gorm:"not null;default:true" json:"active"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

func (f *Feature) BeforeCreate(tx *gorm.DB) error {
	if f.ID == uuid.Nil {
		f.ID = uuid.New()
	}
	return nil
}
