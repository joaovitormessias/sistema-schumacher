package chat

import "testing"

func TestCollectCandidateMediaPrefersImageDataURL(t *testing.T) {
	dataURL := "data:image/jpeg;base64,/9j/2Q=="

	items := collectCandidateMedia([]Message{
		{
			ID:   "msg-1",
			Kind: "IMAGE",
			NormalizedPayload: map[string]interface{}{
				"image_data_url":  dataURL,
				"image_url":       "https://files.example.test/doc.jpg",
				"image_mime_type": "image/jpeg",
			},
		},
	})

	if len(items) != 1 {
		t.Fatalf("expected one media item, got %d", len(items))
	}
	if items[0].URL != dataURL {
		t.Fatalf("unexpected media url: %s", items[0].URL)
	}
	if items[0].MimeType != "image/jpeg" {
		t.Fatalf("unexpected mime type: %s", items[0].MimeType)
	}
	if items[0].MessageID != "msg-1" {
		t.Fatalf("unexpected message id: %s", items[0].MessageID)
	}
}

func TestCollectCandidateMediaFallsBackToImageURL(t *testing.T) {
	items := collectCandidateMedia([]Message{
		{
			ID:   "msg-1",
			Kind: "IMAGE",
			NormalizedPayload: map[string]interface{}{
				"image_url":       "https://files.example.test/doc.jpg",
				"image_mime_type": "image/jpeg",
			},
		},
	})

	if len(items) != 1 {
		t.Fatalf("expected one media item, got %d", len(items))
	}
	if items[0].URL != "https://files.example.test/doc.jpg" {
		t.Fatalf("unexpected media url: %s", items[0].URL)
	}
}
