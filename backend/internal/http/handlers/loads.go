package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/maceo-kwik/drumkit/backend/internal/domain"
	"github.com/maceo-kwik/drumkit/backend/internal/turvo"
)

// LoadHandler exposes HTTP handlers for listing, creating, and fetching loads.
// It delegates remote operations to the Turvo API client and converts between
// Turvo models and the app's domain models via the Mapper.
type LoadHandler struct {
	TurvoClient *turvo.Client
	TurvoMapper *turvo.Mapper
}

// NewLoadHandler returns a fully wired LoadHandler instance.
func NewLoadHandler(client *turvo.Client, mapper *turvo.Mapper) *LoadHandler {
	return &LoadHandler{
		TurvoClient: client,
		TurvoMapper: mapper,
	}
}

// RegisterRoutes mounts all load-related endpoints under /api/loads and
// also exposes /api/customers for a minimal customer list used by the UI.
func (h *LoadHandler) RegisterRoutes(r *chi.Mux) {
	r.Route("/api/loads", func(r chi.Router) {
		r.Get("/", h.ListLoads)
		r.Post("/", h.CreateLoad)
		r.Get("/{id}", h.GetLoadByID)
		r.Put("/{id}", h.UpdateLoad) // Stretch goal
	})
	r.Get("/api/customers", h.ListCustomers)
}

// ListLoads returns a paged list of loads. Query parameters are whitelisted
// and forwarded to Turvo (e.g. start, pageSize, created[gte], status[eq], sortBy).
func (h *LoadHandler) ListLoads(w http.ResponseWriter, r *http.Request) {
	log.Printf("ListLoads called")
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
	// filters (basic only)
	for _, key := range []string{
		"createdDate[gte]", "lastUpdatedOn[lte]",
		"created[gte]", "updated[lte]",
		"customId[eq]", "status[eq]", "sortBy",
	} {
		if v := q.Get(key); v != "" {
			forward.Set(key, v)
		}
	}
	// global search: q -> customId[like]
	if v := q.Get("q"); v != "" {
		forward.Set("customId[like]", v)
	}
	// Default: created in last 90 days (Turvo may restrict large windows)
	if forward.Get("created[gte]") == "" && forward.Get("createdDate[gte]") == "" {
		forward.Set("createdDate[gte]", time.Now().AddDate(0, 0, -90).UTC().Format(time.RFC3339))
	}
	// Sensible default page size
	if forward.Get("pageSize") == "" {
		forward.Set("pageSize", "24")
	}

	log.Printf("About to call ListShipmentsPageWithQuery")
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
	// Fetch full details for each shipment to obtain lane (pickup/destination)
	type idxShipment struct {
		idx int
		s   turvo.Shipment
	}
	enriched := make([]turvo.Shipment, len(shipments))
	copy(enriched, shipments)
	sem := make(chan struct{}, 6)
	pending := 0
	for _, s := range shipments {
		if s.Lane != nil && (s.Lane.Start != "" || s.Lane.End != "") {
			continue // already has lane; no need to enrich
		}
		pending++
	}
	if pending > 0 {
		results := make(chan idxShipment, pending)
		for i, s := range shipments {
			if s.Lane != nil && (s.Lane.Start != "" || s.Lane.End != "") {
				continue
			}
			sem <- struct{}{}
			go func(i int, id int) {
				defer func() { <-sem }()
				ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
				defer cancel()
				detail, err := h.TurvoClient.GetShipment(ctx, strconv.Itoa(id))
				if err != nil || detail == nil {
					results <- idxShipment{idx: i, s: shipments[i]}
					return
				}
				results <- idxShipment{idx: i, s: *detail}
			}(i, s.ID)
		}
		for k := 0; k < pending; k++ {
			res := <-results
			enriched[res.idx] = res.s
		}
	}
	var loads []*domain.Load
	for _, s := range enriched {
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

// CreateLoad creates a shipment in Turvo based on the posted Load payload.
// On success, it returns the mapped Load of the created shipment.
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

// GetLoadByID fetches a single shipment by Turvo id and maps it into a Load.
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

func (h *LoadHandler) UpdateLoad(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "update load", "id": id})
}

// ListCustomers proxies a minimal list of customers from Turvo for dropdowns.
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
