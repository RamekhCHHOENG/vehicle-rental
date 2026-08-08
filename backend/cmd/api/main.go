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
	cfg := config.Load()

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("database: %v", err)
	}

	auth := &handlers.AuthHandler{DB: db, JWTSecret: cfg.JWTSecret}
	vehicles := &handlers.VehicleHandler{DB: db}
	admin := &handlers.AdminHandler{DB: db}

	r := chi.NewRouter()
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type"},
		AllowCredentials: true,
	}))

	r.Get("/api/health", func(w http.ResponseWriter, r *http.Request) {
		httputil.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

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
	})

	// Uploaded photos served as static files.
	r.Handle("/uploads/*", http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads"))))

	log.Printf("API listening on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatal(err)
	}
}
