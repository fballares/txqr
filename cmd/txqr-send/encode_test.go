package main

import (
	"strings"
	"testing"

	"github.com/divan/txqr"
)

func TestEncodeTransferShortIsStatic(t *testing.T) {
	chunks, used, err := encodeTransfer("hello clipboard", txqr.DefaultClipboardProfile())
	if err != nil {
		t.Fatal(err)
	}
	if len(chunks) != 1 {
		t.Fatalf("expected single frame for short text, got %d", len(chunks))
	}
	if used.ChunkLen < len("hello clipboard") {
		t.Fatalf("chunk %d too small for payload", used.ChunkLen)
	}
}

func TestEncodeTransferLongUsesFountain(t *testing.T) {
	b := make([]byte, 800)
	for i := range b {
		b[i] = 'a'
	}
	text := string(b)

	chunks, used, err := encodeTransfer(text, txqr.ProfileForPayload(len(text)))
	if err != nil {
		t.Fatal(err)
	}
	if len(chunks) < 2 {
		t.Fatalf("expected multiple fountain frames, got %d", len(chunks))
	}
	if used.FPS < 1 || used.FPS > 10 {
		t.Fatalf("unexpected fps %d", used.FPS)
	}
	svg, err := renderSVG(chunks[0])
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(svg, "crispEdges") || !strings.Contains(svg, "#000000") {
		t.Fatalf("unexpected SVG: %s", svg[:min(120, len(svg))])
	}
}

func TestParseHotkey(t *testing.T) {
	mods, key, err := parseHotkey("Ctrl+Shift+Q")
	if err != nil {
		t.Fatal(err)
	}
	if len(mods) != 2 {
		t.Fatalf("mods: %v", mods)
	}
	if key == 0 {
		t.Fatal("key unset")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
