package payments

import "testing"

func TestParseProviderData(t *testing.T) {
	raw := []byte(`{"data":{"url":"https://checkout.example/pix","pixQrCode":"000201abc"}}`)
	parsed, checkoutURL, pixCode := parseProviderData(raw)

	if parsed == nil {
		t.Fatalf("expected parsed provider payload")
	}
	if checkoutURL == nil || *checkoutURL != "https://checkout.example/pix" {
		t.Fatalf("unexpected checkout url: %#v", checkoutURL)
	}
	if pixCode == nil || *pixCode != "000201abc" {
		t.Fatalf("unexpected pix code: %#v", pixCode)
	}
}
