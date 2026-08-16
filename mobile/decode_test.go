package txqr

import (
	"strings"
	"testing"
	"time"

	"github.com/divan/txqr"
)

func TestDecode(t *testing.T) {
	payload := strings.Repeat("hello world ", 50)
	enc := txqr.NewEncoder(40)
	chunks, err := enc.Encode(payload)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	dec := NewDecoder()
	for _, chunk := range chunks {
		if err := dec.Decode(chunk); err != nil {
			t.Fatalf("Decode failed: %v", err)
		}
		if dec.IsCompleted() {
			break
		}
	}
	if !dec.IsCompleted() {
		t.Fatal("expected decode to complete")
	}
	if got := dec.Data(); got != payload {
		t.Fatalf("expected %q, got %q", payload, got)
	}
	if dec.Progress() != 100 {
		t.Fatalf("expected progress 100, got %d", dec.Progress())
	}
	if dec.Length() != len(payload) {
		t.Fatalf("expected length %d, got %d", len(payload), dec.Length())
	}
}

func TestDecodeIgnoresInvalidAndCompleted(t *testing.T) {
	payload := "txqr-windows-text"
	enc := txqr.NewEncoder(8)
	chunks, err := enc.Encode(payload)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	dec := NewDecoder()
	if err := dec.Decode("not-a-frame"); err == nil {
		t.Fatal("expected validation error for invalid frame")
	}

	for _, chunk := range chunks {
		if err := dec.Decode(chunk); err != nil {
			t.Fatalf("Decode failed: %v", err)
		}
	}
	if !dec.IsCompleted() {
		t.Fatal("expected decode to complete")
	}

	// Further frames should be ignored once complete.
	if err := dec.Decode(chunks[0]); err != nil {
		t.Fatalf("unexpected error after completion: %v", err)
	}
	if got := dec.Data(); got != payload {
		t.Fatalf("expected %q, got %q", payload, got)
	}
}

func TestProgressIncreases(t *testing.T) {
	payload := strings.Repeat("abcdefghij", 30)
	enc := txqr.NewEncoder(20)
	chunks, err := enc.Encode(payload)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	dec := NewDecoder()
	prev := -1
	for _, chunk := range chunks {
		if err := dec.Decode(chunk); err != nil {
			t.Fatalf("Decode failed: %v", err)
		}
		p := dec.Progress()
		if p < prev {
			t.Fatalf("progress decreased from %d to %d", prev, p)
		}
		prev = p
		if dec.IsCompleted() {
			break
		}
	}
	if prev != 100 {
		t.Fatalf("expected final progress 100, got %d", prev)
	}
}

func TestTotalTime(t *testing.T) {
	dur := 12345678 * time.Microsecond // 12.345678s
	got := formatDuration(dur)
	expected := "12.3s"
	if got != expected {
		t.Fatalf("Expected str to be '%s', but got '%s'", expected, got)
	}
}
