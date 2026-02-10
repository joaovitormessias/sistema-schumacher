package service_orders

import "time"

// Tipos de OS conforme documento de negócio
type ServiceOrderType string

const (
	Preventive ServiceOrderType = "PREVENTIVE"
	Corrective ServiceOrderType = "CORRECTIVE"
)

type ServiceOrderStatus string

const (
	StatusOpen       ServiceOrderStatus = "OPEN"
	StatusInProgress ServiceOrderStatus = "IN_PROGRESS"
	StatusClosed     ServiceOrderStatus = "CLOSED"
	StatusCancelled  ServiceOrderStatus = "CANCELLED"
)

type ServiceOrder struct {
	ID               string             `json:"id"`
	OrderNumber      int                `json:"order_number"`
	BusID            string             `json:"bus_id"`
	BusPlate         *string            `json:"bus_plate,omitempty"`
	DriverID         *string            `json:"driver_id"`
	DriverName       *string            `json:"driver_name,omitempty"`
	OrderType        ServiceOrderType   `json:"order_type"`
	Status           ServiceOrderStatus `json:"status"`
	Description      string             `json:"description"`
	OdometerKm       *int               `json:"odometer_km"`
	ScheduledDate    *time.Time         `json:"scheduled_date"`
	Location         string             `json:"location"`
	OpenedAt         time.Time          `json:"opened_at"`
	ClosedAt         *time.Time         `json:"closed_at"`
	ClosedOdometerKm *int               `json:"closed_odometer_km"`
	NextPreventiveKm *int               `json:"next_preventive_km"`
	Notes            *string            `json:"notes"`
	CreatedBy        *string            `json:"created_by"`
	CreatedAt        time.Time          `json:"created_at"`
	UpdatedAt        time.Time          `json:"updated_at"`
}

type CreateServiceOrderInput struct {
	BusID         string           `json:"bus_id"`
	DriverID      *string          `json:"driver_id"`
	OrderType     ServiceOrderType `json:"order_type"`
	Description   string           `json:"description"`
	OdometerKm    *int             `json:"odometer_km"`
	ScheduledDate *time.Time       `json:"scheduled_date"`
	Location      *string          `json:"location"`
	Notes         *string          `json:"notes"`
}

type UpdateServiceOrderInput struct {
	DriverID         *string `json:"driver_id"`
	Description      *string `json:"description"`
	OdometerKm       *int    `json:"odometer_km"`
	Location         *string `json:"location"`
	Notes            *string `json:"notes"`
	NextPreventiveKm *int    `json:"next_preventive_km"`
}

type CloseServiceOrderInput struct {
	ClosedOdometerKm *int `json:"closed_odometer_km"`
	NextPreventiveKm *int `json:"next_preventive_km"`
}

type ListFilter struct {
	Limit     int
	Offset    int
	Status    *ServiceOrderStatus
	OrderType *ServiceOrderType
	BusID     *string
}
