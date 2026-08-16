package main

import (
	"fmt"

	"github.com/divan/txqr"
	"github.com/divan/txqr/qr"
)

// encodeTransfer builds QR frames for text using a clipboard-tuned profile.
// If a frame exceeds QR capacity it retries with a smaller chunk size.
func encodeTransfer(text string, p txqr.ClipboardProfile) (chunks []string, used txqr.ClipboardProfile, err error) {
	used = p
	if used.ChunkLen <= 0 {
		used = txqr.ProfileForPayload(len(text))
	}
	if used.FPS <= 0 {
		used.FPS = 6
	}
	if used.QRSize <= 0 {
		used.QRSize = 560
	}
	if used.Redundancy <= 0 {
		used.Redundancy = 2.0
	}

	chunk := used.ChunkLen
	for chunk >= 40 {
		enc := txqr.NewEncoder(chunk)
		enc.SetRedundancyFactor(used.Redundancy)
		frames, encErr := enc.Encode(text)
		if encErr != nil {
			return nil, used, encErr
		}
		// Validate capacity with the same high-contrast encoder used for display.
		if _, qrErr := qr.EncodeSVG(frames[0], qr.Medium); qrErr != nil {
			chunk = chunk * 3 / 4
			continue
		}
		used.ChunkLen = chunk
		return frames, used, nil
	}
	return nil, used, fmt.Errorf("text too large to fit into QR frames (try less text)")
}

// renderSVG returns a crisp, scalable vector QR for overlay display.
func renderSVG(chunk string) (string, error) {
	return qr.EncodeSVG(chunk, qr.Medium)
}
