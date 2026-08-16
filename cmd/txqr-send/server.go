package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/gif"
	"image/png"
	"net/http"
	"strings"

	"github.com/divan/txqr"
	"github.com/divan/txqr/qr"
)

// SenderServer serves the paste UI and encodes text into animated QR frames.
type SenderServer struct {
	ChunkLen int
	FPS      int
	QRSize   int
	Initial  string
}

type encodeRequest struct {
	Text     string  `json:"text"`
	ChunkLen int     `json:"chunk_len"`
	FPS      int     `json:"fps"`
	QRSize   int     `json:"qr_size"`
	Format   string  `json:"format"` // "gif" or "frames"
	Redundancy float64 `json:"redundancy"`
}

type encodeResponse struct {
	FrameCount int      `json:"frame_count"`
	FPS        int      `json:"fps"`
	Bytes      int      `json:"bytes"`
	GIF        string   `json:"gif,omitempty"`    // data URL
	Frames     []string `json:"frames,omitempty"` // PNG data URLs
	Error      string   `json:"error,omitempty"`
}

func (s *SenderServer) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, indexHTML, htmlEscape(s.Initial), s.ChunkLen, s.FPS, s.QRSize)
}

func (s *SenderServer) handleEncode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	var req encodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, encodeResponse{Error: "invalid JSON body"})
		return
	}

	text := req.Text
	if strings.TrimSpace(text) == "" {
		writeJSON(w, http.StatusBadRequest, encodeResponse{Error: "text is empty"})
		return
	}

	chunkLen := req.ChunkLen
	if chunkLen <= 0 {
		chunkLen = s.ChunkLen
	}
	fps := req.FPS
	if fps <= 0 {
		fps = s.FPS
	}
	qrSize := req.QRSize
	if qrSize <= 0 {
		qrSize = s.QRSize
	}
	redundancy := req.Redundancy
	if redundancy <= 0 {
		redundancy = 2.0
	}

	enc := txqr.NewEncoder(chunkLen)
	enc.SetRedundancyFactor(redundancy)
	chunks, err := enc.Encode(text)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, encodeResponse{Error: err.Error()})
		return
	}

	resp := encodeResponse{
		FrameCount: len(chunks),
		FPS:        fps,
		Bytes:      len(text),
	}

	format := strings.ToLower(req.Format)
	if format == "" {
		format = "gif"
	}

	switch format {
	case "frames":
		frames := make([]string, 0, len(chunks))
		for _, chunk := range chunks {
			img, err := qr.Encode(chunk, qrSize, qr.Medium)
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, encodeResponse{Error: err.Error()})
				return
			}
			dataURL, err := pngDataURL(img)
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, encodeResponse{Error: err.Error()})
				return
			}
			frames = append(frames, dataURL)
		}
		resp.Frames = frames
	default:
		gifBytes, err := animatedGIF(chunks, qrSize, fps)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, encodeResponse{Error: err.Error()})
			return
		}
		resp.GIF = "data:image/gif;base64," + base64.StdEncoding.EncodeToString(gifBytes)
	}

	writeJSON(w, http.StatusOK, resp)
}

func animatedGIF(chunks []string, qrSize, fps int) ([]byte, error) {
	out := &gif.GIF{
		Image: make([]*image.Paletted, len(chunks)),
		Delay: make([]int, len(chunks)),
	}
	delay := 100 / fps
	if delay < 1 {
		delay = 1
	}
	for i, chunk := range chunks {
		img, err := qr.Encode(chunk, qrSize, qr.Medium)
		if err != nil {
			return nil, fmt.Errorf("QR encode: %w", err)
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

func pngDataURL(img image.Image) (string, error) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", err
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func htmlEscape(s string) string {
	r := strings.NewReplacer(
		`&`, "&amp;",
		`<`, "&lt;",
		`>`, "&gt;",
		`"`, "&quot;",
		`'`, "&#39;",
	)
	return r.Replace(s)
}
