package bookings

import "testing"

func TestNormalizePassengersUsesPluralListWhenPresent(t *testing.T) {
	passengers := normalizePassengers(
		PassengerInput{Name: "Ignorado"},
		[]PassengerInput{
			{Name: " Maria ", Document: " 123 ", Phone: " 4899 ", Email: " maria@example.com "},
			{Name: "Joao"},
		},
	)

	if len(passengers) != 2 {
		t.Fatalf("expected 2 passengers, got %d", len(passengers))
	}
	if passengers[0].Name != "Maria" || passengers[0].Document != "123" || passengers[0].Phone != "4899" || passengers[0].Email != "maria@example.com" {
		t.Fatalf("unexpected normalization result: %+v", passengers[0])
	}
	if passengers[1].Name != "Joao" {
		t.Fatalf("unexpected second passenger: %+v", passengers[1])
	}
}

func TestNormalizePassengersFallsBackToSingularPassenger(t *testing.T) {
	passengers := normalizePassengers(
		PassengerInput{Name: " Ana ", Document: " 456 "},
		nil,
	)

	if len(passengers) != 1 {
		t.Fatalf("expected 1 passenger, got %d", len(passengers))
	}
	if passengers[0].Name != "Ana" || passengers[0].Document != "456" {
		t.Fatalf("unexpected normalization result: %+v", passengers[0])
	}
}

func TestNormalizePassengersPreservesExplicitItemsForValidation(t *testing.T) {
	passengers := normalizePassengers(
		PassengerInput{},
		[]PassengerInput{
			{},
			{Name: "Carlos"},
		},
	)

	if len(passengers) != 2 {
		t.Fatalf("expected 2 passengers, got %d", len(passengers))
	}
	if passengers[0].Name != "" {
		t.Fatalf("expected first passenger to remain empty for validation, got %+v", passengers[0])
	}
	if passengers[1].Name != "Carlos" {
		t.Fatalf("unexpected second passenger result: %+v", passengers[1])
	}
}

func TestBookingDetailsWithPassengersKeepsFirstPassengerAlias(t *testing.T) {
	details := bookingDetailsWithPassengers(Booking{ID: "BK-1"}, []BookingPassenger{
		{ID: "PS-1", Name: "Maria"},
		{ID: "PS-2", Name: "Joao"},
	})

	if len(details.Passengers) != 2 {
		t.Fatalf("expected 2 passengers in response, got %d", len(details.Passengers))
	}
	if details.Passenger.ID != "PS-1" {
		t.Fatalf("expected first passenger alias, got %+v", details.Passenger)
	}
}
