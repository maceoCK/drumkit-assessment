package domain

import "time"

// Load represents the Drumkit Load object structure.
type Load struct {
	ExternalTMSLoadID string         `json:"externalTMSLoadID"`
	FreightLoadID     string         `json:"freightLoadID,omitempty"`
	Status            string         `json:"status"`
	Customer          Party          `json:"customer"`
	BillTo            *Party         `json:"billTo,omitempty"`
	Pickup            Stop           `json:"pickup"`
	Consignee         Stop           `json:"consignee"`
	Carrier           *Carrier       `json:"carrier,omitempty"`
	Specifications    Specifications `json:"specifications"`
	InPalletCount     int            `json:"inPalletCount,omitempty"`
	TotalWeight       float64        `json:"totalWeight,omitempty"`
	// Add other fields from the Drumkit Load schema as needed
}

// Party represents a customer or bill-to party.
type Party struct {
	ExternalTMSId string `json:"externalTMSId,omitempty"`
	Name          string `json:"name"`
	AddressLine1  string `json:"addressLine1"`
	AddressLine2  string `json:"addressLine2,omitempty"`
	City          string `json:"city"`
	State         string `json:"state"`
	Zipcode       string `json:"zipcode"`
	Country       string `json:"country"`
	Contact       string `json:"contact,omitempty"`
	Phone         string `json:"phone,omitempty"`
	Email         string `json:"email,omitempty"`
	RefNumber     string `json:"refNumber,omitempty"`
}

// Stop represents a pickup or consignee location.
type Stop struct {
	Name         string    `json:"name"`
	AddressLine1 string    `json:"addressLine1"`
	AddressLine2 string    `json:"addressLine2,omitempty"`
	City         string    `json:"city"`
	State        string    `json:"state"`
	Zipcode      string    `json:"zipcode"`
	Country      string    `json:"country"`
	ReadyTime    time.Time `json:"readyTime,omitempty"`
	ApptTime     time.Time `json:"apptTime,omitempty"`
	MustDeliver  time.Time `json:"mustDeliver,omitempty"`
	Timezone     string    `json:"timezone,omitempty"`
	WarehouseId  string    `json:"warehouseId,omitempty"`
	Contact      string    `json:"contact,omitempty"`
	Notes        string    `json:"notes,omitempty"`
}

// Carrier represents the carrier information.
type Carrier struct {
	MCNumber    string `json:"mcNumber,omitempty"`
	DOTNumber   string `json:"dotNumber,omitempty"`
	Name        string `json:"name,omitempty"`
	SCAC        string `json:"scac,omitempty"`
	Dispatcher  string `json:"dispatcher,omitempty"`
	DriverName  string `json:"driverName,omitempty"`
	DriverPhone string `json:"driverPhone,omitempty"`
	TruckID     string `json:"truckId,omitempty"`
	TrailerID   string `json:"trailerId,omitempty"`
}

// RateData holds pricing and rate information.
type RateData struct {
	CustomerRate    float64 `json:"customerRate,omitempty"`
	CarrierRate     float64 `json:"carrierRate,omitempty"`
	LinehaulRate    float64 `json:"linehaulRate,omitempty"`
	FuelSurcharge   float64 `json:"fuelSurcharge,omitempty"`
	DetentionHours  float64 `json:"detentionHours,omitempty"`
	MaxRate         float64 `json:"maxRate,omitempty"`
	ProjectedProfit float64 `json:"projectedProfit,omitempty"`
}

// Specifications defines special requirements for the load.
type Specifications struct {
	Hazmat           bool   `json:"hazmat,omitempty"`
	LiftgatePickup   bool   `json:"liftgatePickup,omitempty"`
	LiftgateDelivery bool   `json:"liftgateDelivery,omitempty"`
	InsideDelivery   bool   `json:"insideDelivery,omitempty"`
	Oversize         bool   `json:"oversize,omitempty"`
	Tarps            bool   `json:"tarps,omitempty"`
	Permits          bool   `json:"permits,omitempty"`
	Escorts          bool   `json:"escorts,omitempty"`
	TempMin          int    `json:"tempMin,omitempty"`
	TempMax          int    `json:"tempMax,omitempty"`
	TempUnit         string `json:"tempUnit,omitempty"`
}
