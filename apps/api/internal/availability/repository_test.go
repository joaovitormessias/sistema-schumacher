package availability

import "testing"

func TestActiveTripStatusClauseAcceptsLegacyActiveStatus(t *testing.T) {
	got := activeTripStatusClause("t.status")
	want := "upper(coalesce(t.status, '')) in ('SCHEDULED', 'IN_PROGRESS', 'ATIVO', 'ACTIVE')"
	if got != want {
		t.Fatalf("activeTripStatusClause() = %q, want %q", got, want)
	}
}

func TestNormalizedTripStatusSQLMapsLegacyStatuses(t *testing.T) {
	got := normalizedTripStatusSQL("t.status")
	want := `case
      when upper(coalesce(t.status, '')) in ('ATIVO', 'ACTIVE') then 'SCHEDULED'
      else upper(coalesce(t.status, ''))
    end`
	if got != want {
		t.Fatalf("normalizedTripStatusSQL() = %q, want %q", got, want)
	}
}
