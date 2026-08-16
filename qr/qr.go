package qr

import (
	"fmt"
	"image"
	"image/color"
	"strings"

	"github.com/makiuchi-d/gozxing"
	zqrcode "github.com/makiuchi-d/gozxing/qrcode"
	"github.com/skip2/go-qrcode"
)

// RecoveryLevel represents QR encoding error detection/recovery capacity.
type RecoveryLevel int

const (
	// Low is Level L: 7% error recovery.
	Low RecoveryLevel = iota

	// Medium is Level M: 15% error recovery. Good default choice.
	Medium

	// High is Level Q: 25% error recovery.
	High

	// Highers is Level H: 30% error recovery.
	Highest
)

// newCode builds a high-contrast QR (pure black modules on pure white)
// with the library quiet zone intact for phone-camera readability.
func newCode(data string, lvl RecoveryLevel) (*qrcode.QRCode, error) {
	code, err := qrcode.New(data, qrcode.RecoveryLevel(lvl))
	if err != nil {
		return nil, fmt.Errorf("encode QR: %v", err)
	}
	code.ForegroundColor = color.Black
	code.BackgroundColor = color.White
	return code, nil
}

// Encode encodes data into a raster QR image (black on white, quiet zone kept).
func Encode(data string, size int, lvl RecoveryLevel) (image.Image, error) {
	code, err := newCode(data, lvl)
	if err != nil {
		return nil, err
	}
	return code.Image(size), nil
}

// EncodeFit encodes data at lvl. Returns an error if the payload
// exceeds QR version limits so callers can shrink chunk size and retry.
func EncodeFit(data string, size int, lvl RecoveryLevel) (image.Image, error) {
	return Encode(data, size, lvl)
}

// EncodeSVG returns a resolution-independent SVG QR code.
// Pure #000 on #FFF, crispEdges, viewBox-scaled — ideal for a resizable
// Windows overlay read by an iPhone camera.
func EncodeSVG(data string, lvl RecoveryLevel) (string, error) {
	code, err := newCode(data, lvl)
	if err != nil {
		return "", err
	}
	bits := code.Bitmap()
	if len(bits) == 0 || len(bits[0]) == 0 {
		return "", fmt.Errorf("empty QR bitmap")
	}
	n := len(bits)

	var path strings.Builder
	// Merge black modules into a compact path (1×1 squares).
	for y := 0; y < n; y++ {
		row := bits[y]
		for x := 0; x < len(row); x++ {
			if !row[x] {
				continue
			}
			// M x,y h1v1h-1z — crisp module square
			fmt.Fprintf(&path, "M%d %dh1v1h-1z", x, y)
		}
	}

	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	fmt.Fprintf(&b,
		`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" width="100%%" height="100%%" shape-rendering="crispEdges" role="img" aria-label="TXQR">`,
		n, n)
	// Full white canvas (includes quiet zone already present in Bitmap).
	fmt.Fprintf(&b, `<rect width="%d" height="%d" fill="#ffffff"/>`, n, n)
	b.WriteString(`<path fill="#000000" d="`)
	b.WriteString(path.String())
	b.WriteString(`"/></svg>`)
	return b.String(), nil
}

// Decode an image with QR code.
func Decode(img image.Image) (string, error) {
	bitmap, err := gozxing.NewBinaryBitmapFromImage(img)
	if err != nil {
		return "", fmt.Errorf("gozxing bitmap: %v", err)
	}

	hints := make(map[gozxing.DecodeHintType]interface{})
	hints[gozxing.DecodeHintType_PURE_BARCODE] = true
	result, err := zqrcode.NewQRCodeReader().Decode(bitmap, hints)
	if err != nil {
		return "", fmt.Errorf("gozxing: %v", err)
	}
	return result.GetText(), nil
}

// String implements Stringer interface.
func (r RecoveryLevel) String() string {
	switch r {
	case Low:
		return "Low"
	case Medium:
		return "Medium"
	case High:
		return "High"
	case Highest:
		return "Highest"
	default:
		return "N/A"
	}
}
