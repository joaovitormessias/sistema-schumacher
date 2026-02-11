package routes

import (
	"net/http/httptest"
	"testing"
)

func TestParseListFilterRoutesValid(t *testing.T) {
	req := httptest.NewRequest("GET", "/routes?limit=10&offset=20&status=active", nil)
	filter, err := parseListFilter(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if filter.Limit != 10 || filter.Offset != 20 || filter.Status != "active" {
		t.Fatalf("unexpected filter: %+v", filter)
	}
}

func TestParseListFilterRoutesInvalidLimit(t *testing.T) {
	req := httptest.NewRequest("GET", "/routes?limit=0", nil)
	if _, err := parseListFilter(req); err == nil {
		t.Fatalf("expected error for invalid limit")
	}
}

func TestParseListFilterRoutesInvalidOffset(t *testing.T) {
	req := httptest.NewRequest("GET", "/routes?offset=-1", nil)
	if _, err := parseListFilter(req); err == nil {
		t.Fatalf("expected error for invalid offset")
	}
}

func TestParseListFilterRoutesInvalidStatus(t *testing.T) {
	req := httptest.NewRequest("GET", "/routes?status=broken", nil)
	if _, err := parseListFilter(req); err == nil {
		t.Fatalf("expected error for invalid status")
	}
}

func TestNormalizeRequirementRule(t *testing.T) {
	tests := []struct {
		name        string
		requirement string
		want        string
	}{
		{
			name:        "Known Rule",
			requirement: "at least two stops are required",
			want:        "at_least_two_stops_required",
		},
		{
			name:        "Known Rule Trim Case",
			requirement: "  ETA_OFFSET_MINUTES must be >= 0 ",
			want:        "eta_offset_minutes_non_negative",
		},
		{
			name:        "Fallback Rule",
			requirement: "custom rule > x",
			want:        "custom_rule_gt_x",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeRequirementRule(tt.requirement)
			if got != tt.want {
				t.Fatalf("normalizeRequirementRule(%q) = %q, want %q", tt.requirement, got, tt.want)
			}
		})
	}
}
