package trip_validations

import "time"

type TripValidation struct {
  ID                 string     `json:"id"`
  TripID             string     `json:"trip_id"`
  OdometerInitial    *int       `json:"odometer_initial"`
  OdometerFinal      *int       `json:"odometer_final"`
  DistanceKM         *int       `json:"distance_km"`
  PassengersExpected int        `json:"passengers_expected"`
  PassengersBoarded  int        `json:"passengers_boarded"`
  PassengersNoShow   int        `json:"passengers_no_show"`
  ValidationNotes    *string    `json:"validation_notes"`
  ValidatedBy        *string    `json:"validated_by"`
  ValidatedAt        *time.Time `json:"validated_at"`
  CreatedAt          time.Time  `json:"created_at"`
  UpdatedAt          time.Time  `json:"updated_at"`
}

type CreateTripValidationInput struct {
  TripID             string  `json:"trip_id"`
  OdometerInitial    *int    `json:"odometer_initial"`
  OdometerFinal      *int    `json:"odometer_final"`
  PassengersExpected int     `json:"passengers_expected"`
  PassengersBoarded  int     `json:"passengers_boarded"`
  PassengersNoShow   int     `json:"passengers_no_show"`
  ValidationNotes    *string `json:"validation_notes"`
}

type UpdateTripValidationInput struct {
  OdometerInitial    *int    `json:"odometer_initial"`
  OdometerFinal      *int    `json:"odometer_final"`
  PassengersExpected *int    `json:"passengers_expected"`
  PassengersBoarded  *int    `json:"passengers_boarded"`
  PassengersNoShow   *int    `json:"passengers_no_show"`
  ValidationNotes    *string `json:"validation_notes"`
}

type ListFilter struct {
  TripID string
  Limit  int
  Offset int
}
