package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	chcors "github.com/go-chi/cors"
	"github.com/maceo-kwik/drumkit/backend/internal/config"
	"github.com/maceo-kwik/drumkit/backend/internal/http/handlers"
	"github.com/maceo-kwik/drumkit/backend/internal/turvo"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create a new Turvo client
	turvoClient, err := turvo.NewClient(cfg)
	if err != nil {
		log.Fatalf("Failed to create Turvo client: %v", err)
	}

	// Create a new mapper
	turvoMapper := turvo.NewMapper(cfg)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(chcors.Handler(chcors.Options{
		AllowedOrigins:   cfg.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health checks
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	r.Get("/readyz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// API routes
	loadHandler := handlers.NewLoadHandler(turvoClient, turvoMapper)
	loadHandler.RegisterRoutes(r)

	log.Println("Server starting on port 8080...")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
