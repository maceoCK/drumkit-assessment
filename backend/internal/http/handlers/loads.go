package handlers

import (
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/maceo-kwik/drumkit/backend/internal/domain"
	"github.com/maceo-kwik/drumkit/backend/internal/turvo"
)

// LoadHandler handles the HTTP requests for loads.
type LoadHandler struct {
	TurvoClient *turvo.Client
	TurvoMapper *turvo.Mapper
}

// NewLoadHandler creates a new LoadHandler.
func NewLoadHandler(client *turvo.Client, mapper *turvo.Mapper) *LoadHandler {
	return &LoadHandler{
		TurvoClient: client,
		TurvoMapper: mapper,
	}
}

// RegisterRoutes registers the load routes to the chi router.
func (h *LoadHandler) RegisterRoutes(r *chi.Mux) {
	r.Route("/api/loads", func(r chi.Router) {
		r.Get("/", h.ListLoads)
		r.Post("/", h.CreateLoad)
		r.Get("/{id}", h.GetLoadByID)
		r.Get("/by-external/{externalTMSLoadID}", h.GetLoadByExternalID)
		r.Put("/{id}", h.UpdateLoad) // Stretch goal
	})
	r.Get("/api/customers", h.ListCustomers)
}

func (h *LoadHandler) ListLoads(w http.ResponseWriter, r *http.Request) {
	// Build query for Turvo with whitelist
	forward := url.Values{}
	q := r.URL.Query()
	// pagination
	if v := q.Get("start"); v != "" {
		forward.Set("start", v)
	}
	if v := q.Get("pageSize"); v != "" {
		forward.Set("pageSize", v)
	}
	// filters
	for _, key := range []string{
		"created[gte]", "updated[lte]", "customId[eq]", "status[eq]", "status[in]",
		"locationId[eq]", "pickupDate[gte]", "pickupDate[lte]", "deliveryDate[gte]", "deliveryDate[lte]",
		"customerId[eq]", "poNumber[eq]", "bolNumber[eq]", "containerNumber[eq]", "proNumber[eq]",
		"routeNumber[eq]", "other[eq]", "truckNumber[eq]", "parentAccount[eq]", "parentAccount[in]",
		"trackingProvider[in]", "serviceAreaKey[eq]", "serviceAreaKey[in]", "sortBy",
	} {
		if v := q.Get(key); v != "" {
			forward.Set(key, v)
		}
	}
	// Default: created in last 90 days (Turvo may restrict large windows)
	if forward.Get("created[gte]") == "" {
		forward.Set("created[gte]", time.Now().AddDate(0, 0, -90).UTC().Format(time.RFC3339))
	}
	// Sensible default page size
	if forward.Get("pageSize") == "" {
		forward.Set("pageSize", "24")
	}

	shipments, meta, err := h.TurvoClient.ListShipmentsPageWithQuery(r.Context(), forward)
	if err != nil {
		if rl, ok := err.(turvo.RateLimitedError); ok {
			if rl.RetryAfter > 0 {
				w.Header().Set("Retry-After", rl.RetryAfter.Round(time.Second).String())
			}
			http.Error(w, rl.Error(), http.StatusTooManyRequests)
			return
		}
		http.Error(w, "turvo list error: "+err.Error(), http.StatusBadGateway)
		return
	}
	var loads []*domain.Load
	for _, s := range shipments {
		l, _ := h.TurvoMapper.FromTurvoShipment(s)
		loads = append(loads, l)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"items": loads,
		"pagination": map[string]any{
			"start":              meta.Start,
			"pageSize":           meta.PageSize,
			"totalRecordsInPage": meta.TotalRecordsInPage,
			"moreAvailable":      meta.MoreAvailable,
		},
	})
}

func (h *LoadHandler) CreateLoad(w http.ResponseWriter, r *http.Request) {
	var load domain.Load
	if err := json.NewDecoder(r.Body).Decode(&load); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}
	shipment, err := h.TurvoMapper.ToTurvoShipment(&load)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	created, err := h.TurvoClient.CreateShipment(r.Context(), shipment)
	if err != nil {
		http.Error(w, "turvo create error: "+err.Error(), http.StatusBadGateway)
		return
	}
	l, _ := h.TurvoMapper.FromTurvoShipment(*created)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(l)
}

func (h *LoadHandler) GetLoadByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	s, err := h.TurvoClient.GetShipment(r.Context(), id)
	if err != nil {
		http.Error(w, "turvo get error: "+err.Error(), http.StatusBadGateway)
		return
	}
	l, _ := h.TurvoMapper.FromTurvoShipment(*s)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(l)
}

func (h *LoadHandler) GetLoadByExternalID(w http.ResponseWriter, r *http.Request) {
	externalID := chi.URLParam(r, "externalTMSLoadID")
	s, err := h.TurvoClient.FindShipmentByExternalID(r.Context(), externalID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	l, _ := h.TurvoMapper.FromTurvoShipment(*s)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(l)
}

func (h *LoadHandler) UpdateLoad(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement logic to update a load in Turvo
	id := chi.URLParam(r, "id")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "update load", "id": id})
}

// ListCustomers proxies Turvo customers list (minimal fields)
func (h *LoadHandler) ListCustomers(w http.ResponseWriter, r *http.Request) {
	forward := url.Values{}
	q := r.URL.Query()
	for _, key := range []string{"start", "pageSize", "name[eq]", "status[eq]", "updated[lte]", "created[gte]"} {
		if v := q.Get(key); v != "" {
			forward.Set(key, v)
		}
	}
	customers, err := h.TurvoClient.ListCustomers(r.Context(), forward)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"items": customers})
}
