package trip_operations

import "testing"

func TestIsAllowedTransitionSequence(t *testing.T) {
	if !isAllowedTransition(TripOperationalStatusRequested, TripOperationalStatusPassengersReady) {
		t.Fatalf("expected REQUESTED -> PASSENGERS_READY to be allowed")
	}
	if !isAllowedTransition(TripOperationalStatusSettled, TripOperationalStatusClosed) {
		t.Fatalf("expected SETTLED -> CLOSED to be allowed")
	}
}

func TestIsAllowedTransitionRejectsSkip(t *testing.T) {
	if isAllowedTransition(TripOperationalStatusRequested, TripOperationalStatusAuthorized) {
		t.Fatalf("expected REQUESTED -> AUTHORIZED to be blocked")
	}
	if isAllowedTransition(TripOperationalStatusInProgress, TripOperationalStatusSettled) {
		t.Fatalf("expected IN_PROGRESS -> SETTLED to be blocked")
	}
}
