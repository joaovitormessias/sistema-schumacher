package trip_operations

import (
	"encoding/json"
	"time"
)

const (
	TripOperationalStatusRequested         = "REQUESTED"
	TripOperationalStatusPassengersReady   = "PASSENGERS_READY"
	TripOperationalStatusItineraryReady    = "ITINERARY_READY"
	TripOperationalStatusDispatchValidated = "DISPATCH_VALIDATED"
	TripOperationalStatusAuthorized        = "AUTHORIZED"
	TripOperationalStatusInProgress        = "IN_PROGRESS"
	TripOperationalStatusReturned          = "RETURNED"
	TripOperationalStatusReturnChecked     = "RETURN_CHECKED"
	TripOperationalStatusSettled           = "SETTLED"
	TripOperationalStatusClosed            = "CLOSED"
)

var ValidOperationalStatuses = map[string]struct{}{
	TripOperationalStatusRequested:         {},
	TripOperationalStatusPassengersReady:   {},
	TripOperationalStatusItineraryReady:    {},
	TripOperationalStatusDispatchValidated: {},
	TripOperationalStatusAuthorized:        {},
	TripOperationalStatusInProgress:        {},
	TripOperationalStatusReturned:          {},
	TripOperationalStatusReturnChecked:     {},
	TripOperationalStatusSettled:           {},
	TripOperationalStatusClosed:            {},
}

type TripRequest struct {
	ID                   string     `json:"id"`
	RouteID              *string    `json:"route_id"`
	Source               string     `json:"source"`
	Status               string     `json:"status"`
	RequesterName        *string    `json:"requester_name"`
	RequesterContact     *string    `json:"requester_contact"`
	RequestedDepartureAt *time.Time `json:"requested_departure_at"`
	Notes                *string    `json:"notes"`
	CreatedBy            *string    `json:"created_by"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

type CreateTripRequestInput struct {
	RouteID              *string    `json:"route_id"`
	Source               string     `json:"source"`
	Status               *string    `json:"status"`
	RequesterName        *string    `json:"requester_name"`
	RequesterContact     *string    `json:"requester_contact"`
	RequestedDepartureAt *time.Time `json:"requested_departure_at"`
	Notes                *string    `json:"notes"`
}

type TripManifestEntry struct {
	ID                 string    `json:"id"`
	TripID             string    `json:"trip_id"`
	BookingPassengerID *string   `json:"booking_passenger_id"`
	PassengerName      string    `json:"passenger_name"`
	PassengerDocument  *string   `json:"passenger_document"`
	PassengerPhone     *string   `json:"passenger_phone"`
	Source             string    `json:"source"`
	Status             string    `json:"status"`
	SeatNumber         *int      `json:"seat_number"`
	IsActive           bool      `json:"is_active"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type CreateManifestEntryInput struct {
	BookingPassengerID *string `json:"booking_passenger_id"`
	PassengerName      string  `json:"passenger_name"`
	PassengerDocument  *string `json:"passenger_document"`
	PassengerPhone     *string `json:"passenger_phone"`
	Status             *string `json:"status"`
	SeatNumber         *int    `json:"seat_number"`
}

type UpdateManifestEntryInput struct {
	PassengerName     *string `json:"passenger_name"`
	PassengerDocument *string `json:"passenger_document"`
	PassengerPhone    *string `json:"passenger_phone"`
	Status            *string `json:"status"`
	SeatNumber        *int    `json:"seat_number"`
	IsActive          *bool   `json:"is_active"`
}

type TripAuthorization struct {
	ID                    string     `json:"id"`
	TripID                string     `json:"trip_id"`
	Authority             string     `json:"authority"`
	Status                string     `json:"status"`
	ProtocolNumber        *string    `json:"protocol_number"`
	LicenseNumber         *string    `json:"license_number"`
	IssuedAt              *time.Time `json:"issued_at"`
	ValidUntil            *time.Time `json:"valid_until"`
	SRCPolicyNumber       *string    `json:"src_policy_number"`
	SRCValidUntil         *time.Time `json:"src_valid_until"`
	ExceptionalDeadlineOK bool       `json:"exceptional_deadline_ok"`
	AttachmentID          *string    `json:"attachment_id"`
	Notes                 *string    `json:"notes"`
	CreatedBy             *string    `json:"created_by"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}

type CreateTripAuthorizationInput struct {
	Authority             string     `json:"authority"`
	Status                string     `json:"status"`
	ProtocolNumber        *string    `json:"protocol_number"`
	LicenseNumber         *string    `json:"license_number"`
	IssuedAt              *time.Time `json:"issued_at"`
	ValidUntil            *time.Time `json:"valid_until"`
	SRCPolicyNumber       *string    `json:"src_policy_number"`
	SRCValidUntil         *time.Time `json:"src_valid_until"`
	ExceptionalDeadlineOK *bool      `json:"exceptional_deadline_ok"`
	AttachmentID          *string    `json:"attachment_id"`
	Notes                 *string    `json:"notes"`
}

type UpdateTripAuthorizationInput struct {
	Authority             *string    `json:"authority"`
	Status                *string    `json:"status"`
	ProtocolNumber        *string    `json:"protocol_number"`
	LicenseNumber         *string    `json:"license_number"`
	IssuedAt              *time.Time `json:"issued_at"`
	ValidUntil            *time.Time `json:"valid_until"`
	SRCPolicyNumber       *string    `json:"src_policy_number"`
	SRCValidUntil         *time.Time `json:"src_valid_until"`
	ExceptionalDeadlineOK *bool      `json:"exceptional_deadline_ok"`
	AttachmentID          *string    `json:"attachment_id"`
	Notes                 *string    `json:"notes"`
}

type TripChecklist struct {
	ID                string          `json:"id"`
	TripID            string          `json:"trip_id"`
	Stage             string          `json:"stage"`
	ChecklistData     json.RawMessage `json:"checklist_data"`
	IsComplete        bool            `json:"is_complete"`
	DocumentsChecked  bool            `json:"documents_checked"`
	TachographChecked bool            `json:"tachograph_checked"`
	ReceiptsChecked   bool            `json:"receipts_checked"`
	RestComplianceOK  bool            `json:"rest_compliance_ok"`
	CompletedBy       *string         `json:"completed_by"`
	CompletedAt       *time.Time      `json:"completed_at"`
	Notes             *string         `json:"notes"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
}

type UpsertChecklistInput struct {
	ChecklistData     json.RawMessage `json:"checklist_data"`
	IsComplete        bool            `json:"is_complete"`
	DocumentsChecked  bool            `json:"documents_checked"`
	TachographChecked bool            `json:"tachograph_checked"`
	ReceiptsChecked   bool            `json:"receipts_checked"`
	RestComplianceOK  *bool           `json:"rest_compliance_ok"`
	Notes             *string         `json:"notes"`
}

type TripDriverReport struct {
	ID             string    `json:"id"`
	TripID         string    `json:"trip_id"`
	DriverID       *string   `json:"driver_id"`
	OdometerStart  *int      `json:"odometer_start"`
	OdometerEnd    *int      `json:"odometer_end"`
	FuelUsedLiters *float64  `json:"fuel_used_liters"`
	Incidents      *string   `json:"incidents"`
	Delays         *string   `json:"delays"`
	RestHours      *float64  `json:"rest_hours"`
	Notes          *string   `json:"notes"`
	SubmittedBy    *string   `json:"submitted_by"`
	SubmittedAt    time.Time `json:"submitted_at"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type UpsertDriverReportInput struct {
	DriverID       *string  `json:"driver_id"`
	OdometerStart  *int     `json:"odometer_start"`
	OdometerEnd    *int     `json:"odometer_end"`
	FuelUsedLiters *float64 `json:"fuel_used_liters"`
	Incidents      *string  `json:"incidents"`
	Delays         *string  `json:"delays"`
	RestHours      *float64 `json:"rest_hours"`
	Notes          *string  `json:"notes"`
}

type TripReceiptReconciliation struct {
	ID                    string     `json:"id"`
	TripID                string     `json:"trip_id"`
	TotalReceiptsAmount   float64    `json:"total_receipts_amount"`
	TotalApprovedExpenses float64    `json:"total_approved_expenses"`
	Difference            float64    `json:"difference"`
	ReceiptsValidated     bool       `json:"receipts_validated"`
	VerifiedExpenseIDs    []string   `json:"verified_expense_ids"`
	Notes                 *string    `json:"notes"`
	ReconciledBy          *string    `json:"reconciled_by"`
	ReconciledAt          *time.Time `json:"reconciled_at"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}

type UpsertReceiptReconciliationInput struct {
	TotalReceiptsAmount float64  `json:"total_receipts_amount"`
	ReceiptsValidated   bool     `json:"receipts_validated"`
	VerifiedExpenseIDs  []string `json:"verified_expense_ids"`
	Notes               *string  `json:"notes"`
}

type TripAttachment struct {
	ID             string          `json:"id"`
	TripID         string          `json:"trip_id"`
	AttachmentType string          `json:"attachment_type"`
	StorageBucket  string          `json:"storage_bucket"`
	StoragePath    string          `json:"storage_path"`
	FileName       string          `json:"file_name"`
	MimeType       *string         `json:"mime_type"`
	FileSize       *int64          `json:"file_size"`
	Metadata       json.RawMessage `json:"metadata"`
	UploadedBy     *string         `json:"uploaded_by"`
	UploadedAt     time.Time       `json:"uploaded_at"`
	CreatedAt      time.Time       `json:"created_at"`
}

type CreateAttachmentInput struct {
	AttachmentType string          `json:"attachment_type"`
	StorageBucket  *string         `json:"storage_bucket"`
	StoragePath    string          `json:"storage_path"`
	FileName       string          `json:"file_name"`
	MimeType       *string         `json:"mime_type"`
	FileSize       *int64          `json:"file_size"`
	Metadata       json.RawMessage `json:"metadata"`
}

type WorkflowAdvanceInput struct {
	ToStatus string `json:"to_status"`
}

type WorkflowBlockedResponse struct {
	Code                string   `json:"code"`
	Message             string   `json:"message"`
	RequirementsMissing []string `json:"requirements_missing"`
}

type OperationalTripState struct {
	ID                  string     `json:"id"`
	Status              string     `json:"status"`
	OperationalStatus   string     `json:"operational_status"`
	DepartureAt         time.Time  `json:"departure_at"`
	EstimatedKM         float64    `json:"estimated_km"`
	DispatchValidatedAt *time.Time `json:"dispatch_validated_at"`
	DispatchValidatedBy *string    `json:"dispatch_validated_by"`
}

type WorkflowValidation struct {
	Allowed             bool
	Code                string
	Message             string
	RequirementsMissing []string
}
