package turvo

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/maceo-kwik/drumkit/backend/internal/config"
	"github.com/maceo-kwik/drumkit/backend/internal/domain"
)

// Mapper handles the mapping between Drumkit and Turvo models.
type Mapper struct {
	cfg *config.Config
}

// NewMapper creates a new Mapper.
func NewMapper(cfg *config.Config) *Mapper {
	return &Mapper{cfg: cfg}
}

// ToTurvoShipment converts a Drumkit Load into a Turvo Shipment.
func (m *Mapper) ToTurvoShipment(load *domain.Load) (Shipment, error) {
	now := time.Now()
	var pickupAt time.Time
	if load.Pickup.ReadyTime != nil && !load.Pickup.ReadyTime.IsZero() {
		pickupAt = *load.Pickup.ReadyTime
	} else {
		pickupAt = now
	}
	var deliveryAt time.Time
	if load.Consignee.MustDeliver != nil && !load.Consignee.MustDeliver.IsZero() {
		deliveryAt = *load.Consignee.MustDeliver
	} else {
		deliveryAt = pickupAt.Add(24 * time.Hour)
	}

	// Default location IDs (unused when SkipDistanceCalculation and lane are provided)
	_ = m.cfg.TurvoDefaultOriginLocationID
	_ = m.cfg.TurvoDefaultDestinationLocationID
	custID := m.cfg.TurvoDefaultCustomerID
	if load.Customer.TurvoID > 0 {
		custID = load.Customer.TurvoID
	}

	// Build lane strings in "city, state" format as required by Turvo
	startLane := strings.TrimSpace(strings.Join([]string{
		strings.TrimSpace(load.Pickup.City),
		strings.TrimSpace(load.Pickup.State),
	}, ", "))
	endLane := strings.TrimSpace(strings.Join([]string{
		strings.TrimSpace(load.Consignee.City),
		strings.TrimSpace(load.Consignee.State),
	}, ", "))

	// Build customer order with nested customer id
	co := CustomerOrder{
		Customer: &struct {
			ID   int    `json:"id"`
			Name string `json:"name,omitempty"`
		}{ID: custID},
		CustomerOrderSourceID: 1,
	}

	shipment := Shipment{
		CustomID:                load.ExternalTMSLoadID,
		LtlShipment:             false,
		StartDate:               DateWithTZ{Date: pickupAt, TimeZone: "UTC"},
		EndDate:                 DateWithTZ{Date: deliveryAt, TimeZone: "UTC"},
		CustomerOrder:           []CustomerOrder{co},
		Lane:                    &Lane{Start: startLane, End: endLane},
		SkipDistanceCalculation: true,
		GlobalRoute:             nil,
	}
	return shipment, nil
}

// FromTurvoShipment converts a Turvo Shipment into a Drumkit Load.
func (m *Mapper) FromTurvoShipment(s Shipment) (*domain.Load, error) {
	// Try to parse Status.Code.Value if present; otherwise leave empty
	status := ""
	if len(s.Status) > 0 {
		var st struct {
			Code struct {
				Value string `json:"value"`
			} `json:"code"`
			Notes string `json:"notes"`
		}
		if err := json.Unmarshal(s.Status, &st); err == nil {
			status = strings.TrimSpace(st.Code.Value)
		}
	}
	if status == "" {
		status = "Unknown"
	}

	customerName := ""
	if len(s.CustomerOrder) > 0 && s.CustomerOrder[0].Customer != nil {
		customerName = s.CustomerOrder[0].Customer.Name
	}

	load := &domain.Load{
		ExternalTMSLoadID: s.CustomID,
		Status:            status,
		CreatedAt:         s.CreatedDate,
		Customer:          domain.Party{Name: customerName},
		Specifications:    &domain.Specifications{},
	}
	return load, nil
}
