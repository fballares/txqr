package qr

import (
	"fmt"
	"image"

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

// Encode encodes data into the image with QR code.
// Uses the library defaults (black modules on white) which scan
// reliably from a bright monitor into a phone camera. The built-in
// quiet zone is kept — phones need that margin in a popup window.
func Encode(data string, size int, lvl RecoveryLevel) (image.Image, error) {
	code, err := qrcode.New(data, qrcode.RecoveryLevel(lvl))
	if err != nil {
		return nil, fmt.Errorf("encode QR: %v", err)
	}
	return code.Image(size), nil
}

// EncodeFit encodes data at lvl. Returns an error if the payload
// exceeds QR version limits so callers can shrink chunk size and retry.
func EncodeFit(data string, size int, lvl RecoveryLevel) (image.Image, error) {
	return Encode(data, size, lvl)
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
