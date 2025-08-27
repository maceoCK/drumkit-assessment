package domain

import "time"

type Load struct {
	ExternalTMSLoadID string          `json:"externalTMSLoadID"`
	FreightLoadID     string          `json:"freightLoadID,omitempty"`
	Status            string          `json:"status"`
	CreatedAt         *time.Time      `json:"createdAt,omitempty"`
	Customer          Party           `json:"customer"`
	BillTo            *Party          `json:"billTo,omitempty"`
	Pickup            Stop            `json:"pickup"`
	Consignee         Stop            `json:"consignee"`
	Carrier           *Carrier        `json:"carrier,omitempty"`
	RateData          *RateData       `json:"rateData,omitempty"`
	Specifications    *Specifications `json:"specifications,omitempty"`
	// Additional derived fields for UI display
	Phase              string   `json:"phase,omitempty"`
	Mode               string   `json:"mode,omitempty"`
	ServiceType        string   `json:"serviceType,omitempty"`
	Services           []string `json:"services,omitempty"`
	Equipment          []string `json:"equipment,omitempty"`
	CustomerTotalMiles *float64 `json:"customerTotalMiles,omitempty"`
	MarginAmount       *float64 `json:"marginAmount,omitempty"`
	MarginValue        *float64 `json:"marginValue,omitempty"`
}

type Party struct {
	TurvoID       int    `json:"turvoId,omitempty"`
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

type Stop struct {
	ExternalTMSId string     `json:"externalTMSId,omitempty"`
	Name          string     `json:"name"`
	AddressLine1  string     `json:"addressLine1"`
	AddressLine2  string     `json:"addressLine2,omitempty"`
	City          string     `json:"city"`
	State         string     `json:"state"`
	Zipcode       string     `json:"zipcode"`
	Country       string     `json:"country"`
	Contact       string     `json:"contact,omitempty"`
	Phone         string     `json:"phone,omitempty"`
	Email         string     `json:"email,omitempty"`
	BusinessHours string     `json:"businessHours,omitempty"`
	RefNumber     string     `json:"refNumber,omitempty"`
	ReadyTime     *time.Time `json:"readyTime,omitempty"`   // pickup
	MustDeliver   *time.Time `json:"mustDeliver,omitempty"` // consignee
	ApptTime      *time.Time `json:"apptTime,omitempty"`
	ApptNote      string     `json:"apptNote,omitempty"`
	Timezone      string     `json:"timezone,omitempty"`
	WarehouseId   string     `json:"warehouseId,omitempty"`
}

type Carrier struct {
	MCNumber                 string     `json:"mcNumber,omitempty"`
	DOTNumber                string     `json:"dotNumber,omitempty"`
	Name                     string     `json:"name,omitempty"`
	Phone                    string     `json:"phone,omitempty"`
	Dispatcher               string     `json:"dispatcher,omitempty"`
	SealNumber               string     `json:"sealNumber,omitempty"`
	SCAC                     string     `json:"scac,omitempty"`
	FirstDriverName          string     `json:"firstDriverName,omitempty"`
	FirstDriverPhone         string     `json:"firstDriverPhone,omitempty"`
	SecondDriverName         string     `json:"secondDriverName,omitempty"`
	SecondDriverPhone        string     `json:"secondDriverPhone,omitempty"`
	Email                    string     `json:"email,omitempty"`
	DispatchCity             string     `json:"dispatchCity,omitempty"`
	DispatchState            string     `json:"dispatchState,omitempty"`
	ExternalTMSTruckId       string     `json:"externalTMSTruckId,omitempty"`
	ExternalTMSTrailerId     string     `json:"externalTMSTrailerId,omitempty"`
	ConfirmationSentTime     *time.Time `json:"confirmationSentTime,omitempty"`
	ConfirmationReceivedTime *time.Time `json:"confirmationReceivedTime,omitempty"`
	DispatchedTime           *time.Time `json:"dispatchedTime,omitempty"`
	ExpectedPickupTime       *time.Time `json:"expectedPickupTime,omitempty"`
	PickupStart              *time.Time `json:"pickupStart,omitempty"`
	PickupEnd                *time.Time `json:"pickupEnd,omitempty"`
	ExpectedDeliveryTime     *time.Time `json:"expectedDeliveryTime,omitempty"`
	DeliveryStart            *time.Time `json:"deliveryStart,omitempty"`
	DeliveryEnd              *time.Time `json:"deliveryEnd,omitempty"`
	SignedBy                 string     `json:"signedBy,omitempty"`
	ExternalTMSId            string     `json:"externalTMSId,omitempty"`
}

type RateData struct {
	CustomerRateType  string  `json:"customerRateType,omitempty"`
	CustomerNumHours  float64 `json:"customerNumHours,omitempty"`
	CustomerLhRateUsd float64 `json:"customerLhRateUsd,omitempty"`
	FscPercent        float64 `json:"fscPercent,omitempty"`
	FscPerMile        float64 `json:"fscPerMile,omitempty"`
	CarrierRateType   string  `json:"carrierRateType,omitempty"`
	CarrierNumHours   float64 `json:"carrierNumHours,omitempty"`
	CarrierLhRateUsd  float64 `json:"carrierLhRateUsd,omitempty"`
	CarrierMaxRate    float64 `json:"carrierMaxRate,omitempty"`
	NetProfitUsd      float64 `json:"netProfitUsd,omitempty"`
	ProfitPercent     float64 `json:"profitPercent,omitempty"`
}

type Specifications struct {
	MinTempFahrenheit float64 `json:"minTempFahrenheit,omitempty"`
	MaxTempFahrenheit float64 `json:"maxTempFahrenheit,omitempty"`
	LiftgatePickup    bool    `json:"liftgatePickup,omitempty"`
	LiftgateDelivery  bool    `json:"liftgateDelivery,omitempty"`
	InsidePickup      bool    `json:"insidePickup,omitempty"`
	InsideDelivery    bool    `json:"insideDelivery,omitempty"`
	Tarps             bool    `json:"tarps,omitempty"`
	Oversized         bool    `json:"oversized,omitempty"`
	Hazmat            bool    `json:"hazmat,omitempty"`
	Straps            bool    `json:"straps,omitempty"`
	Permits           bool    `json:"permits,omitempty"`
	Escorts           bool    `json:"escorts,omitempty"`
	Seal              bool    `json:"seal,omitempty"`
	CustomBonded      bool    `json:"customBonded,omitempty"`
	Labor             bool    `json:"labor,omitempty"`
	InPalletCount     int     `json:"inPalletCount,omitempty"`
	OutPalletCount    int     `json:"outPalletCount,omitempty"`
	NumCommodities    int     `json:"numCommodities,omitempty"`
	TotalWeight       float64 `json:"totalWeight,omitempty"`
	BillableWeight    float64 `json:"billableWeight,omitempty"`
	PoNums            string  `json:"poNums,omitempty"`
	Operator          string  `json:"operator,omitempty"`
	RouteMiles        float64 `json:"routeMiles,omitempty"`
}
