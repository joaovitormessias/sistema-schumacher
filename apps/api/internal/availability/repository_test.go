package availability

import "testing"

func TestNormalizeSearchTextFoldsAccents(t *testing.T) {
	if got := normalizeSearchText("Concórdia/SC"); got != "concordia/sc" {
		t.Fatalf("expected folded concordia/sc, got %q", got)
	}
	if got := normalizeSearchText("Santa Ines/MA"); got != "santa ines/ma" {
		t.Fatalf("expected santa ines/ma, got %q", got)
	}
	if got := normalizeSearchText("  Chapecó  /SC "); got != "chapeco/sc" {
		t.Fatalf("expected folded chapeco/sc, got %q", got)
	}
}

func TestNormalizedSearchColumnSQLUsesAccentInsensitiveTranslation(t *testing.T) {
	got := normalizedSearchColumnSQL("destination_stop.display_name")
	want := "translate(lower(coalesce(destination_stop.display_name, '')), 'áàâãäéèêëíìîïóòôõöúùûüçñ', 'aaaaaeeeeiiiiooooouuuucn')"
	if got != want {
		t.Fatalf("unexpected sql expression: %q", got)
	}
}
