package routes

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSearchCityCandidatesEmptyResultsReturnsEmptyList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[]`))
	}))
	defer server.Close()

	geocoder := &NominatimGeocoder{
		client:      server.Client(),
		endpointURL: server.URL,
		userAgent:   "test-agent",
		countryCode: "br",
	}

	items, err := geocoder.SearchCityCandidates(context.Background(), "Xyz", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected no candidates, got %d", len(items))
	}
}

func TestGeocodeCityEmptyResultsReturnsNoResultError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[]`))
	}))
	defer server.Close()

	geocoder := &NominatimGeocoder{
		client:      server.Client(),
		endpointURL: server.URL,
		userAgent:   "test-agent",
		countryCode: "br",
	}

	_, _, err := geocoder.GeocodeCity(context.Background(), "CidadeTeste/SC")
	if err == nil {
		t.Fatalf("expected error when geocoding city without matches")
	}
	if err.Error() != "no geocoding result" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStateCodeFromNameSupportsAccents(t *testing.T) {
	tests := []struct {
		state string
		want  string
	}{
		{state: "S\u00E3o Paulo", want: "SP"},
		{state: "Maranh\u00E3o", want: "MA"},
		{state: "Par\u00E1", want: "PA"},
		{state: "Cear\u00E1", want: "CE"},
	}

	for _, tt := range tests {
		got := stateCodeFromName(tt.state)
		if got != tt.want {
			t.Fatalf("stateCodeFromName(%q) = %q, want %q", tt.state, got, tt.want)
		}
	}
}
