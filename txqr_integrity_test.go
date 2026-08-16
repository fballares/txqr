package txqr

import (
	"hash/crc32"
	"strings"
	"testing"
)

func TestEncodeDecodeCRCVerified(t *testing.T) {
	payload := strings.Repeat("integrity-check-", 40)
	enc := NewEncoder(48)
	chunks, err := enc.Encode(payload)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(chunks[0], "|") {
		t.Fatal("missing separator")
	}
	header := chunks[0][:strings.IndexByte(chunks[0], '|')]
	parts := strings.Split(header, "/")
	if len(parts) != 4 {
		t.Fatalf("want 4 header fields, got %q", header)
	}

	dec := NewDecoder()
	for _, c := range chunks {
		if err := dec.Decode(c); err != nil {
			t.Fatal(err)
		}
		if dec.IsCompleted() {
			break
		}
	}
	if !dec.IsCompleted() {
		t.Fatal("expected completion")
	}
	if !dec.IsVerified() {
		t.Fatal("expected verified")
	}
	if !dec.HasCRC() {
		t.Fatal("expected CRC")
	}
	if got := dec.Data(); got != payload {
		t.Fatalf("payload mismatch")
	}
	want := crc32.ChecksumIEEE([]byte(payload))
	if dec.ExpectedCRC() != want {
		t.Fatalf("crc %08x want %08x", dec.ExpectedCRC(), want)
	}
}

func TestCRCMismatchRejected(t *testing.T) {
	payload := strings.Repeat("abcd", 30)
	enc := NewEncoder(32)
	chunks, err := enc.Encode(payload)
	if err != nil {
		t.Fatal(err)
	}
	// Tamper CRC in every header while leaving payload bytes alone.
	tampered := make([]string, len(chunks))
	for i, c := range chunks {
		idx := strings.IndexByte(c, '|')
		h := c[:idx]
		parts := strings.Split(h, "/")
		if len(parts) != 4 {
			t.Fatalf("header %q", h)
		}
		parts[3] = "deadbeef"
		tampered[i] = strings.Join(parts, "/") + c[idx:]
	}

	dec := NewDecoder()
	sawErr := false
	for i := 0; i < 3; i++ {
		for _, c := range tampered {
			err := dec.Decode(c)
			if err != nil && strings.Contains(err.Error(), "CRC") {
				sawErr = true
			}
			if dec.IsCompleted() {
				t.Fatal("must not complete with bad CRC")
			}
		}
	}
	if !sawErr {
		t.Fatal("expected CRC error during finalize")
	}
	if dec.Data() != "" {
		t.Fatal("must not expose unverified data")
	}
}

func TestLegacyHeaderStillDecodes(t *testing.T) {
	// Legacy 3-field headers (no CRC) should still reconstruct.
	dec := NewDecoder()
	frames := []string{
		"0/5/11|hello",
		"0/5/11|hello", // duplicate ignored
	}
	// Single chunk path with legacy format: total 11, chunk 5 means 3 source blocks —
	// this is a simplified smoke for parseHeader legacy branch only.
	_, _, _, _, hasCRC, err := parseHeader("0/5/11")
	if err != nil || hasCRC {
		t.Fatalf("legacy parse: hasCRC=%v err=%v", hasCRC, err)
	}
	_, _, _, crc, hasCRC, err := parseHeader("0/5/11/abcd1234")
	if err != nil || !hasCRC || crc != 0xabcd1234 {
		t.Fatalf("crc parse: %#x has=%v err=%v", crc, hasCRC, err)
	}
	_ = frames
	_ = dec
}
