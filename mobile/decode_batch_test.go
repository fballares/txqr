package txqr

import (
	"strings"
	"testing"

	"github.com/divan/txqr"
)

func TestDecodeBatchMultiQR(t *testing.T) {
	payload := strings.Repeat("multi-qr-", 40)
	enc := txqr.NewEncoder(32)
	chunks, err := enc.Encode(payload)
	if err != nil {
		t.Fatal(err)
	}
	if len(chunks) < 2 {
		t.Fatalf("need multiple chunks, got %d", len(chunks))
	}

	dec := NewDecoder()
	// Simulate one camera frame seeing two QRs at once (plus noise).
	batch := "https://example.com\n" + chunks[0] + "\n" + chunks[len(chunks)/2] + "\nnot-a-qr"
	n := dec.DecodeBatch(batch)
	if n < 1 {
		t.Fatalf("expected accepted frames, got %d", n)
	}
	if dec.UniqueFrames() < 1 {
		t.Fatal("expected unique frames")
	}

	// Feed remaining via further batches until complete.
	for !dec.IsCompleted() {
		for i := 0; i+1 < len(chunks); i += 2 {
			joined := chunks[i] + "\n" + chunks[i+1]
			dec.DecodeBatch(joined)
			if dec.IsCompleted() {
				break
			}
		}
		if !dec.IsCompleted() {
			dec.DecodeBatch(strings.Join(chunks, "\n"))
		}
	}
	if dec.Data() != payload {
		t.Fatalf("payload mismatch")
	}
}
