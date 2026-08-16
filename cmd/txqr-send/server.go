package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/divan/txqr"
)

// SenderServer serves the background popup UI and encode APIs.
type SenderServer struct {
	mu sync.RWMutex

	Default txqr.ClipboardProfile
	Hotkey  string
	Streams int // 0 = auto (pick 1 vs 2 by benefit); 1–4 = force

	latest *transfer
}

type transfer struct {
	Text       string
	Bytes      int
	FrameCount int
	FPS        int
	ChunkLen   int
	Redundancy float64
	Streams    int
	Created    time.Time
	Chunks     []string // raw TXQR payloads (source of truth for SVG render)
	Frames     []string // optional inline SVG for small transfers
	Image      string   // first SVG (preview)
	Static     bool
	Error      string
}

type encodeRequest struct {
	Text       string  `json:"text"`
	ChunkLen   int     `json:"chunk_len"`
	FPS        int     `json:"fps"`
	QRSize     int     `json:"qr_size"`
	Redundancy float64 `json:"redundancy"`
	Streams    int     `json:"streams"`
	Auto       bool    `json:"auto"`
}

type encodeResponse struct {
	FrameCount int      `json:"frame_count"`
	FPS        int      `json:"fps"`
	Bytes      int      `json:"bytes"`
	ChunkLen   int      `json:"chunk_len"`
	Redundancy float64  `json:"redundancy"`
	Streams    int      `json:"streams"`
	Static     bool     `json:"static"`
	Image      string   `json:"image,omitempty"`
	Frames     []string `json:"frames,omitempty"`
	Format     string   `json:"format"` // "svg"
	Looping    bool     `json:"looping"`
	Error      string   `json:"error,omitempty"`
}

func (s *SenderServer) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	p := s.Default
	fmt.Fprintf(w, indexHTML, htmlEscape(s.Hotkey), p.FPS, p.QRSize)
}

func (s *SenderServer) handlePopup(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	streams := 1
	s.mu.RLock()
	if s.latest != nil && s.latest.Streams > 0 {
		streams = s.latest.Streams
	} else if s.Streams > 0 {
		streams = s.Streams
	}
	s.mu.RUnlock()
	fmt.Fprintf(w, popupHTML, htmlEscape(s.Hotkey), streams, streams)
}

func (s *SenderServer) handleLatest(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.latest == nil {
		writeJSON(w, http.StatusOK, encodeResponse{Error: "no transfer yet — copy text and press " + s.Hotkey})
		return
	}
	resp := encodeResponse{
		FrameCount: s.latest.FrameCount,
		FPS:        s.latest.FPS,
		Bytes:      s.latest.Bytes,
		ChunkLen:   s.latest.ChunkLen,
		Redundancy: s.latest.Redundancy,
		Streams:    s.latest.Streams,
		Static:     s.latest.Static,
		Image:      s.latest.Image,
		Format:     "svg",
		Looping:    !s.latest.Static,
		Error:      s.latest.Error,
	}
	// Inline SVG only for small sets; larger transfers use /api/frame?i=
	if r.URL.Query().Get("format") == "frames" && len(s.latest.Frames) > 0 {
		resp.Frames = s.latest.Frames
	}
	writeJSON(w, http.StatusOK, resp)
}

// handleFrame serves one high-contrast SVG QR for overlay animation.
func (s *SenderServer) handleFrame(w http.ResponseWriter, r *http.Request) {
	idx, err := strconv.Atoi(r.URL.Query().Get("i"))
	if err != nil || idx < 0 {
		http.Error(w, "bad index", http.StatusBadRequest)
		return
	}
	s.mu.RLock()
	tr := s.latest
	s.mu.RUnlock()
	if tr == nil || len(tr.Chunks) == 0 {
		http.Error(w, "no transfer", http.StatusNotFound)
		return
	}
	if idx >= len(tr.Chunks) {
		http.Error(w, "index out of range", http.StatusBadRequest)
		return
	}
	svg, err := renderSVG(tr.Chunks[idx])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "image/svg+xml; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write([]byte(svg))
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
	tr, err := s.buildTransfer(req.Text, req)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, encodeResponse{Error: err.Error()})
		return
	}
	s.setLatest(tr)
	writeJSON(w, http.StatusOK, encodeResponse{
		FrameCount: tr.FrameCount,
		FPS:        tr.FPS,
		Bytes:      tr.Bytes,
		ChunkLen:   tr.ChunkLen,
		Redundancy: tr.Redundancy,
		Streams:    tr.Streams,
		Static:     tr.Static,
		Image:      tr.Image,
		Frames:     tr.Frames,
		Format:     "svg",
		Looping:    !tr.Static,
		Error:      tr.Error,
	})
}

func (s *SenderServer) buildTransfer(text string, req encodeRequest) (*transfer, error) {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	if strings.TrimSpace(text) == "" {
		return nil, fmt.Errorf("clipboard is empty — copy text first")
	}

	p := s.Default
	if req.Auto || (req.ChunkLen <= 0 && req.FPS <= 0) {
		p = txqr.ProfileForPayload(len(text))
	}
	if req.ChunkLen > 0 {
		p.ChunkLen = req.ChunkLen
	}
	if req.FPS > 0 {
		p.FPS = req.FPS
	}
	if req.QRSize > 0 {
		p.QRSize = req.QRSize
	}
	if req.Redundancy > 0 {
		p.Redundancy = req.Redundancy
	}

	streams := req.Streams
	if streams <= 0 {
		streams = s.Streams
	}

	chunks, used, err := encodeTransfer(text, p)
	if err != nil {
		return &transfer{Error: err.Error(), Created: time.Now()}, err
	}

	if streams <= 0 {
		streams = txqr.SuggestedStreamsForTransfer(len(text), len(chunks), used.FPS)
	}
	if streams < 1 {
		streams = 1
	}
	if streams > 4 {
		streams = 4
	}
	if len(chunks) == 1 {
		streams = 1
	}

	tr := &transfer{
		Text:       text,
		Bytes:      len(text),
		FrameCount: len(chunks),
		FPS:        used.FPS,
		ChunkLen:   used.ChunkLen,
		Redundancy: used.Redundancy,
		Streams:    streams,
		Created:    time.Now(),
		Chunks:     chunks,
		Static:     len(chunks) == 1,
	}

	// Inline a modest number of SVGs so small transfers start instantly;
	// larger ones fetch /api/frame on demand (still vector, still crisp).
	const inlineLimit = 64
	limit := len(chunks)
	if limit > inlineLimit {
		limit = inlineLimit
	}
	frames := make([]string, 0, limit)
	for i := 0; i < limit; i++ {
		svg, err := renderSVG(chunks[i])
		if err != nil {
			return nil, err
		}
		frames = append(frames, svg)
	}
	tr.Frames = frames
	if len(frames) > 0 {
		tr.Image = frames[0]
	}
	return tr, nil
}

func (s *SenderServer) setLatest(tr *transfer) {
	s.mu.Lock()
	s.latest = tr
	s.mu.Unlock()
}

func (s *SenderServer) showClipboard() error {
	text, err := readClipboardText()
	if err != nil {
		tr := &transfer{Error: "clipboard read failed: " + err.Error(), Created: time.Now()}
		s.setLatest(tr)
		return err
	}
	tr, err := s.buildTransfer(text, encodeRequest{Auto: true})
	if err != nil {
		if tr == nil {
			tr = &transfer{Error: err.Error(), Created: time.Now()}
		}
		s.setLatest(tr)
		return err
	}
	s.setLatest(tr)
	return nil
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
