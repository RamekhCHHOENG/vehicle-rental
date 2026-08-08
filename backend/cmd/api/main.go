package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/ramekhchhoeng/car-rental/backend/internal/config"
	"github.com/ramekhchhoeng/car-rental/backend/internal/database"
	"github.com/ramekhchhoeng/car-rental/backend/internal/handlers"
	"github.com/ramekhchhoeng/car-rental/backend/internal/httputil"
	"github.com/ramekhchhoeng/car-rental/backend/internal/middleware"
	"github.com/ramekhchhoeng/car-rental/backend/internal/models"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("configuration: %v", err)
	}

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("database: %v", err)
	}

	auth := &handlers.AuthHandler{DB: db, JWTSecret: cfg.JWTSecret}
	vehicles := &handlers.VehicleHandler{DB: db}
	admin := &handlers.AdminHandler{DB: db}
	bookings := &handlers.BookingHandler{DB: db}
	metadata := &handlers.MetadataHandler{DB: db}
	adminMetadata := &handlers.AdminMetadataHandler{DB: db}
	metadataImport := &handlers.ImportHandler{DB: db}

	r := chi.NewRouter()
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{cfg.WebOrigin},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type"},
		AllowCredentials: true,
	}))

	r.Get("/api/health", func(w http.ResponseWriter, r *http.Request) {
		httputil.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// The vocabulary listings are written in: provinces, makes, models,
	// features. Public because the browse filters need it before anyone logs in.
	r.Get("/api/metadata", metadata.Get)
	// Models are fetched per make rather than shipped with /api/metadata: a few
	// thousand makes is a small payload, all of their models is not.
	r.Get("/api/makes/{id}/models", metadata.ListModels)

	r.Route("/api/auth", func(r chi.Router) {
		r.Post("/register", auth.Register)
		r.Post("/login", auth.Login)
		r.Post("/logout", auth.Logout)
		r.With(middleware.RequireAuth(cfg.JWTSecret)).Get("/me", auth.Me)
	})

	// Public browse (GetPublic also lets a logged-in owner/admin see non-approved).
	r.Route("/api/vehicles", func(r chi.Router) {
		r.Get("/", vehicles.ListPublic)
		r.With(middleware.OptionalAuth(cfg.JWTSecret)).Get("/{id}", vehicles.GetPublic)
	})

	// Owner-only vehicle management.
	r.Route("/api/owner", func(r chi.Router) {
		r.Use(middleware.RequireAuth(cfg.JWTSecret))
		r.Use(middleware.RequireRole(models.RoleOwner))
		r.Post("/vehicles", vehicles.CreateOwn)
		r.Get("/vehicles", vehicles.ListOwn)
		r.Put("/vehicles/{id}", vehicles.UpdateOwn)
		r.Delete("/vehicles/{id}", vehicles.DeleteOwn)
		r.Post("/vehicles/{id}/photos", vehicles.UploadPhoto)
		r.Delete("/vehicles/{id}/photos/{photoId}", vehicles.DeletePhoto)
		r.Get("/bookings", bookings.ListForOwner)
		r.Post("/bookings/{id}/confirm", bookings.Confirm)
		r.Post("/bookings/{id}/reject", bookings.RejectBooking)
		r.Post("/bookings/{id}/complete", bookings.Complete)
	})

	// Renter bookings.
	r.Route("/api/bookings", func(r chi.Router) {
		r.Use(middleware.RequireAuth(cfg.JWTSecret))
		r.Use(middleware.RequireRole(models.RoleRenter))
		r.Post("/", bookings.Create)
		r.Get("/", bookings.ListOwn)
		r.Post("/{id}/cancel", bookings.Cancel)
		r.Post("/{id}/review", bookings.Review)
	})

	// Admin verification panel.
	r.Route("/api/admin", func(r chi.Router) {
		r.Use(middleware.RequireAuth(cfg.JWTSecret))
		r.Use(middleware.RequireRole(models.RoleAdmin))
		r.Get("/vehicles", admin.ListVehicles)
		r.Post("/vehicles/{id}/approve", admin.Approve)
		r.Post("/vehicles/{id}/reject", admin.Reject)
		r.Get("/users", admin.ListUsers)
		r.Get("/stats", admin.Stats)

		// Managing that vocabulary. Deletes are refused for anything a listing
		// points at — see AdminMetadataHandler.
		r.Get("/metadata", adminMetadata.ListAll)
		// Paged and searched separately: a few thousand makes with their models
		// attached is not a payload, it is a download.
		r.Get("/makes", adminMetadata.ListMakes)
		r.Get("/makes/{id}/models", adminMetadata.ListModelsForMake)
		r.Post("/provinces", adminMetadata.CreateProvince)
		r.Put("/provinces/{id}", adminMetadata.UpdateProvince)
		r.Delete("/provinces/{id}", adminMetadata.DeleteProvince)
		r.Post("/makes", adminMetadata.CreateMake)
		r.Put("/makes/{id}", adminMetadata.UpdateMake)
		r.Delete("/makes/{id}", adminMetadata.DeleteMake)
		r.Post("/models", adminMetadata.CreateModel)
		r.Put("/models/{id}", adminMetadata.UpdateModel)
		r.Delete("/models/{id}", adminMetadata.DeleteModel)
		r.Post("/features", adminMetadata.CreateFeature)
		r.Put("/features/{id}", adminMetadata.UpdateFeature)
		r.Delete("/features/{id}", adminMetadata.DeleteFeature)

		// Bulk import from NHTSA's open vehicle database. Runs in the
		// background — the model half is roughly 1,900 upstream requests.
		r.Get("/metadata/import", metadataImport.Status)
		r.Post("/metadata/import", metadataImport.Start)
	})

	// Uploaded photos served as static files.
	r.Handle("/uploads/*", http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads"))))

	log.Printf("API listening on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatal(err)
	}
}
