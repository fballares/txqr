package main

import (
	"fmt"
	"image"
	"image/gif"
	"image/png"
	"bytes"
	"encoding/base64"

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
		if _, qrErr := qr.Encode(frames[0], used.QRSize, qr.Medium); qrErr != nil {
			chunk = chunk * 3 / 4
			continue
		}
		used.ChunkLen = chunk
		return frames, used, nil
	}
	return nil, used, fmt.Errorf("text too large to fit into QR frames (try less text)")
}

func renderGIF(chunks []string, qrSize, fps int) ([]byte, error) {
	out := &gif.GIF{
		Image: make([]*image.Paletted, len(chunks)),
		Delay: make([]int, len(chunks)),
	}
	delay := 100 / fps
	if delay < 1 {
		delay = 1
	}
	// Minimum ~12 (≈8fps) can be aggressive; clamp floor to 10 (10fps max)
	// only when fps requested is already high — keep caller FPS otherwise.
	for i, chunk := range chunks {
		img, err := qr.Encode(chunk, qrSize, qr.Medium)
		if err != nil {
			return nil, fmt.Errorf("QR encode frame %d: %w", i, err)
		}
		paletted, ok := img.(*image.Paletted)
		if !ok {
			return nil, fmt.Errorf("QR encoder returned non-paletted image")
		}
		out.Image[i] = paletted
		out.Delay[i] = delay
	}
	var buf bytes.Buffer
	if err := gif.EncodeAll(&buf, out); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func renderPNG(chunk string, qrSize int) ([]byte, error) {
	img, err := qr.Encode(chunk, qrSize, qr.Medium)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func dataURL(mime string, raw []byte) string {
	return "data:" + mime + ";base64," + base64.StdEncoding.EncodeToString(raw)
}
