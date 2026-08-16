package qr

import (
	"image"
	"os"
	"strings"
	"testing"
)

func TestDecoder(t *testing.T) {
	filename := "testdata/helloworld_qr.png"
	fd, err := os.Open(filename)
	if err != nil {
		t.Fatalf("Open test file: %v", err)
	}
	img, _, err := image.Decode(fd)
	if err != nil {
		t.Fatalf("Decode test image: %v", err)
	}
	str, err := Decode(img)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if str != "hello, world" {
		t.Fatalf("Expected 'hello, world', but got '%s'", str)
	}
}

func TestEncodeSVGHighContrast(t *testing.T) {
	svg, err := EncodeSVG("0/8/20|hello-svg", Medium)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`viewBox=`,
		`fill="#ffffff"`,
		`fill="#000000"`,
		`shape-rendering="crispEdges"`,
		`<path `,
	} {
		if !strings.Contains(svg, want) {
			t.Fatalf("SVG missing %q (len=%d)", want, len(svg))
		}
	}
	// Must be vector (no raster payload).
	if strings.Contains(svg, "base64") || strings.Contains(svg, "<image") {
		t.Fatal("SVG should not embed raster data")
	}
}

func TestEncodeSVGRoundTripViaRaster(t *testing.T) {
	payload := "1/40/100|abcdef"
	img, err := Encode(payload, 256, Medium)
	if err != nil {
		t.Fatal(err)
	}
	got, err := Decode(img)
	if err != nil {
		t.Fatal(err)
	}
	if got != payload {
		t.Fatalf("got %q want %q", got, payload)
	}
}
