package turvo

import (
	"encoding/json"
	"time"
)

// KeyValuePair represents a simple key/value item as used by Turvo enums.
type KeyValuePair struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// Status represents the status of a Turvo shipment.
type Status struct {
	Code  KeyValuePair `json:"code"`
	Notes string       `json:"notes,omitempty"`
}

// Equipment defines equipment requirements for a shipment.
type Equipment struct {
	Type           KeyValuePair  `json:"type"`
	Weight         *float64      `json:"weight,omitempty"`
	WeightUnits    *KeyValuePair `json:"weightUnits,omitempty"`
	Temp           *float64      `json:"temp,omitempty"`
	TempUnits      *KeyValuePair `json:"tempUnits,omitempty"`
	Size           *KeyValuePair `json:"size,omitempty"`
	Description    *string       `json:"description,omitempty"`
	ShipmentLength *float64      `json:"shipmentLength,omitempty"`
}

// Contributor represents a user associated with a shipment.
type Contributor struct {
	Title           *KeyValuePair    `json:"title,omitempty"`
	ContributorUser *ContributorUser `json:"contributorUser,omitempty"`
}

// ContributorUser holds the ID of a contributor.
type ContributorUser struct {
	ID int `json:"id"`
}

// GlobalRoute represents a stop in the shipment's journey.
type GlobalRoute struct {
	Name                       string            `json:"name,omitempty"`
	AppointmentNo              string            `json:"appointmentNo,omitempty"`
	Locode                     string            `json:"locode,omitempty"`
	SchedulingType             *KeyValuePair     `json:"schedulingType,omitempty"`
	StopType                   KeyValuePair      `json:"stopType"`
	Timezone                   string            `json:"timezone,omitempty"`
	Location                   Location          `json:"location"`
	Sequence                   int               `json:"sequence"`
	SegmentSequence            int               `json:"segmentSequence,omitempty"`
	GlobalShipLocationSourceID string            `json:"globalShipLocationSourceId,omitempty"`
	State                      string            `json:"state,omitempty"`
	Appointment                Appointment       `json:"appointment"`
	PlannedAppointmentDate     *Appointment      `json:"plannedAppointmentDate,omitempty"`
	OriginalAppointmentDate    *Appointment      `json:"originalAppointmentDate,omitempty"`
	PlannedRequestedDate       *Appointment      `json:"plannedRequestedDate,omitempty"`
	FlexAttributes             []FlexAttribute   `json:"flexAttributes,omitempty"`
	ActualPickupDate           *ActualPickupDate `json:"actualPickupDate,omitempty"`
	ExpectedDwellTime          *DwellTime        `json:"expectedDwellTime,omitempty"`
	FragmentDistance           *Distance         `json:"fragmentDistance,omitempty"`
	Distance                   *Distance         `json:"distance,omitempty"`
	Services                   []KeyValuePair    `json:"services,omitempty"`
	PoNumbers                  []string          `json:"poNumbers,omitempty"`
	Notes                      string            `json:"notes,omitempty"`
	RecalculateDistance        bool              `json:"recalculateDistance,omitempty"`
	Contact                    *Contact          `json:"contact,omitempty"`
}

// Location holds the ID of a location.
type Location struct {
	ID int `json:"id"`
}

// Appointment represents scheduling information for a stop.
type Appointment struct {
	Date         time.Time `json:"date"`
	Flex         int       `json:"flex"`
	Timezone     string    `json:"timezone,omitempty"`
	HasTime      bool      `json:"hasTime"`
	Confirmation bool      `json:"appointmentConfirmation,omitempty"`
}

// FlexAttribute is a generic key-value pair for custom attributes.
type FlexAttribute struct {
	Type      KeyValuePair `json:"type"`
	Value     string       `json:"value"`
	Name      string       `json:"name,omitempty"`
	Shareable bool         `json:"shareable,omitempty"`
}

// ActualPickupDate holds arrival and departure times.
type ActualPickupDate struct {
	Arrival  *time.Time `json:"arrival,omitempty"`
	Departed *time.Time `json:"departed,omitempty"`
}

// DwellTime represents a duration of time.
type DwellTime struct {
	Units KeyValuePair `json:"units"`
	Value int          `json:"value"`
}

// Distance represents a distance with units.
type Distance struct {
	Units KeyValuePair `json:"units"`
	Value float64      `json:"value"`
}

// Contact holds the ID of a contact person.
type Contact struct {
	ID int `json:"id"`
}

// Transportation defines the mode and service type.
type Transportation struct {
	Mode        KeyValuePair `json:"mode"`
	ServiceType KeyValuePair `json:"serviceType"`
}

// DateWithTZ matches Turvo format { date, timeZone }.
type DateWithTZ struct {
	Date     time.Time `json:"date"`
	TimeZone string    `json:"timeZone,omitempty"`
}

// Shipment is the top-level object for a Turvo shipment used by the app.
type Shipment struct {
	ID                      int             `json:"id,omitempty"`
	CustomID                string          `json:"customId,omitempty"`
	LtlShipment             bool            `json:"ltlShipment"`
	StartDate               DateWithTZ      `json:"startDate"`
	EndDate                 DateWithTZ      `json:"endDate"`
	CreatedDate             *time.Time      `json:"createdDate,omitempty"`
	Updated                 *time.Time      `json:"updated,omitempty"`
	LastUpdatedOn           *time.Time      `json:"lastUpdatedOn,omitempty"`
	Status                  json.RawMessage `json:"status,omitempty"`
	Equipment               []Equipment     `json:"equipment,omitempty"`
	Contributors            []Contributor   `json:"contributors,omitempty"`
	Lane                    *Lane           `json:"lane,omitempty"`
	GlobalRoute             []GlobalRoute   `json:"globalRoute,omitempty"`
	SkipDistanceCalculation bool            `json:"skipDistanceCalculation,omitempty"`
	ModeInfo                []interface{}   `json:"modeInfo,omitempty"` // Define if structure is known
	FlexAttributes          []FlexAttribute `json:"flexAttributes,omitempty"`
	Groups                  []interface{}   `json:"groups,omitempty"` // Define if structure is known
	CustomerOrder           []CustomerOrder `json:"customerOrder"`
	Margin                  *Margin         `json:"margin,omitempty"`
	Services                []KeyValuePair  `json:"services,omitempty"`
	CarrierOrder            []CarrierOrder  `json:"carrierOrder,omitempty"`
	UseRoutingGuide         bool            `json:"use_routing_guide,omitempty"`
}

// Lane represents the start and end points of a shipment.
type Lane struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

// CustomerOrder links a customer to the shipment (minimal fields for create).
type CustomerOrder struct {
	ID       int  `json:"id,omitempty"`
	Deleted  bool `json:"deleted,omitempty"`
	Customer *struct {
		ID   int    `json:"id"`
		Name string `json:"name,omitempty"`
	} `json:"customer,omitempty"`
	CustomerID            int `json:"customerId,omitempty"`
	CustomerOrderSourceID int `json:"customerOrderSourceId,omitempty"`
}

// CarrierOrder links a carrier to the shipment (kept for completeness).
type CarrierOrder struct {
	CarrierID            int `json:"carrierId"`
	CarrierOrderSourceID int `json:"carrierOrderSourceId"`
}

// Margin represents margin information for a shipment.
type Margin struct {
	MinPay float64 `json:"minPay,omitempty"`
	MaxPay float64 `json:"maxPay,omitempty"`
}

// Order represents a Turvo Order.
type Order struct {
	StartDate                 time.Time       `json:"start_date"`
	EndDate                   time.Time       `json:"end_date"`
	PlannedPickup             *time.Time      `json:"planned_pickup,omitempty"`
	PlannedDelivery           *time.Time      `json:"planned_delivery,omitempty"`
	StartDateType             *KeyValuePair   `json:"start_date_type,omitempty"`
	EndDateType               *KeyValuePair   `json:"end_date_type,omitempty"`
	Customer                  OrderCustomer   `json:"customer"`
	OrderType                 *KeyValuePair   `json:"order_type,omitempty"`
	Direction                 *KeyValuePair   `json:"direction,omitempty"`
	Origin                    interface{}     `json:"origin"`      // Define if structure is known
	Destination               interface{}     `json:"destination"` // Define if structure is known
	OriginFlexAttributes      []FlexAttribute `json:"origin_flex_attributes,omitempty"`
	DestinationFlexAttributes []FlexAttribute `json:"destination_flex_attributes,omitempty"`
	Items                     []OrderItem     `json:"items"`
	ExternalIDs               []ExternalID    `json:"external_ids,omitempty"`
	FlexAttributes            []FlexAttribute `json:"flex_attributes,omitempty"`
	Shipments                 []Shipment      `json:"shipments,omitempty"`
	Carrier                   *OrderCarrier   `json:"carrier,omitempty"`
	UserGroups                []interface{}   `json:"user_groups,omitempty"` // Define if structure is known
}

// OrderCustomer holds the customer ID for an order.
type OrderCustomer struct {
	ID int `json:"id"`
}

// OrderItem represents an item within an order.
type OrderItem struct {
	Ref                  string                    `json:"ref,omitempty"`
	Item                 interface{}               `json:"item"` // Define if structure is known
	Notes                string                    `json:"notes,omitempty"`
	Status               *Status                   `json:"status,omitempty"`
	Quantity             ItemQuantity              `json:"quantity"`
	ActualQuantity       *ItemQuantity             `json:"actualQuantity,omitempty"`
	HandlingQuantity     *ItemQuantity             `json:"handlingQuantity,omitempty"`
	Name                 string                    `json:"name"`
	Attributes           interface{}               `json:"attributes,omitempty"` // Define if structure is known
	Category             ItemCategory              `json:"category"`
	Inventory            interface{}               `json:"inventory,omitempty"` // Define if structure is known
	Dimensions           *ItemDimensions           `json:"dimensions,omitempty"`
	Value                *ItemValue                `json:"value,omitempty"`
	ItemDetails          []interface{}             `json:"item_details,omitempty"` // Define if structure is known
	HazardClass          string                    `json:"hazardClass,omitempty"`
	MaxStackCount        int                       `json:"maxStackCount,omitempty"`
	StackDimensionsLimit *ItemStackDimensionsLimit `json:"stackDimensionsLimit,omitempty"`
	LoadBearingCapacity  *ItemLoadBearingCapacity  `json:"loadBearingCapacity,omitempty"`
	ShippingName         string                    `json:"shippingName,omitempty"`
	Identification       string                    `json:"identification,omitempty"`
	LotNumber            string                    `json:"lot_number,omitempty"`
	FreightClass         *ItemFreightClass         `json:"freight_class,omitempty"`
	NMFC                 string                    `json:"nmfc,omitempty"`
	NMFCSub              string                    `json:"nmfc_sub,omitempty"`
	PackingGroup         *ItemPackingGroup         `json:"packingGroup,omitempty"`
	EmergencyContact     string                    `json:"emergencyContact,omitempty"`
	Volume               *ItemVolume               `json:"volume,omitempty"`
	MinTemp              *ItemTemperature          `json:"minTemp,omitempty"`
	MaxTemp              *ItemTemperature          `json:"maxTemp,omitempty"`
	UnitNetWeight        *ItemWeight               `json:"unit_net_weight,omitempty"`
	UnitGrossWeight      *ItemWeight               `json:"unit_gross_weight,omitempty"`
	Costs                *OrderCosts               `json:"costs,omitempty"`
}

// ItemQuantity represents the quantity of an item.
type ItemQuantity struct {
	// Define if structure is known
}

// ItemCategory represents the category of an item.
type ItemCategory struct {
	// Define if structure is known
}

// ItemDimensions represents the dimensions of an item.
type ItemDimensions struct {
	// Define if structure is known
}

// ItemValue represents the value of an item.
type ItemValue struct {
	// Define if structure is known
}

// ItemStackDimensionsLimit represents the stack dimensions limit of an item.
type ItemStackDimensionsLimit struct {
	// Define if structure is known
}

// ItemLoadBearingCapacity represents the load-bearing capacity of an item.
type ItemLoadBearingCapacity struct {
	// Define if structure is known
}

// ItemFreightClass represents the freight class of an item.
type ItemFreightClass struct {
	// Define if structure is known
}

// ItemPackingGroup represents the packing group of an item.
type ItemPackingGroup struct {
	// Define if structure is known
}

// ItemVolume represents the volume of an item.
type ItemVolume struct {
	// Define if structure is known
}

// ItemTemperature represents the temperature of an item.
type ItemTemperature struct {
	// Define if structure is known
}

// ItemWeight represents the weight of an item.
type ItemWeight struct {
	// Define if structure is known
}

// OrderCosts represents the costs of an order.
type OrderCosts struct {
	// Define if structure is known
}

// ExternalID represents an external identifier for an order.
type ExternalID struct {
	// Define if structure is known
}

// OrderCarrier represents the carrier of an order.
type OrderCarrier struct {
	// Define if structure is known
}
