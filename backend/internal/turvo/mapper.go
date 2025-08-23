package turvo

import (
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
	pickupAt := load.Pickup.ReadyTime
	if pickupAt.IsZero() {
		pickupAt = now
	}
	deliveryAt := load.Consignee.MustDeliver
	if deliveryAt.IsZero() {
		deliveryAt = pickupAt.Add(24 * time.Hour)
	}

	originID := m.cfg.TurvoDefaultOriginLocationID
	destID := m.cfg.TurvoDefaultDestinationLocationID
	custID := m.cfg.TurvoDefaultCustomerID

	shipment := Shipment{
		CustomID:      load.ExternalTMSLoadID,
		LtlShipment:   false,
		StartDate:     pickupAt,
		EndDate:       deliveryAt,
		CustomerOrder: []CustomerOrder{{CustomerID: custID, CustomerOrderSourceID: 0}},
		GlobalRoute: []GlobalRoute{
			{
				StopType:    KeyValuePair{Key: "PICKUP", Value: "PICKUP"},
				Location:    Location{ID: originID},
				Sequence:    1,
				Appointment: Appointment{Date: pickupAt, Flex: 0, HasTime: true},
			},
			{
				StopType:    KeyValuePair{Key: "DROPOFF", Value: "DROPOFF"},
				Location:    Location{ID: destID},
				Sequence:    2,
				Appointment: Appointment{Date: deliveryAt, Flex: 0, HasTime: true},
			},
		},
	}
	return shipment, nil
}

// FromTurvoShipment converts a Turvo Shipment into a Drumkit Load.
func (m *Mapper) FromTurvoShipment(s Shipment) (*domain.Load, error) {
	// Show Turvo's status text directly for user clarity
	status := ""
	if s.Status != nil {
		status = strings.TrimSpace(s.Status.Code.Value)
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
		Customer:          domain.Party{Name: customerName},
		Specifications:    domain.Specifications{},
	}
	return load, nil
}
