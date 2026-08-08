package database

import (
	"log"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/ramekhchhoeng/car-rental/backend/internal/config"
	"github.com/ramekhchhoeng/car-rental/backend/internal/models"
)

// Connect opens the Postgres connection, runs migrations, and seeds the admin user.
func Connect(cfg config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.DBURL), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(
		&models.User{},
		&models.Vehicle{},
		&models.VehiclePhoto{},
		&models.Booking{},
		&models.Review{},
	); err != nil {
		return nil, err
	}

	if err := seedAdmin(db, cfg); err != nil {
		return nil, err
	}

	return db, nil
}

func seedAdmin(db *gorm.DB, cfg config.Config) error {
	var count int64
	if err := db.Model(&models.User{}).Where("role = ?", models.RoleAdmin).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(cfg.AdminPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	admin := models.User{
		Email:        cfg.AdminEmail,
		PasswordHash: string(hash),
		FullName:     "Admin",
		Role:         models.RoleAdmin,
	}
	if err := db.Create(&admin).Error; err != nil {
		return err
	}
	log.Printf("seeded admin user: %s", cfg.AdminEmail)
	return nil
}
