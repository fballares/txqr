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
	Streams int // 0 = auto; 1–4 = force

	latest *transfer
	paused bool
}

type transfer struct {
	Text       string
	Bytes      int
	FrameCount int
	FPS        int
	ChunkLen   int
	Redundancy float64
	Streams    int
	CRCHex     string
	ETASeconds float64
	Created    time.Time
	Chunks     []string
	Frames     []string
	Image      string
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
	CRCHex     string   `json:"crc_hex,omitempty"`
	ETASeconds float64  `json:"eta_seconds"`
	Image      string   `json:"image,omitempty"`
	Frames     []string `json:"frames,omitempty"`
	Format     string   `json:"format"`
	Looping    bool     `json:"looping"`
	Paused     bool     `json:"paused"`
	Error      string   `json:"error,omitempty"`
}

type statusResponse struct {
	Hotkey      string  `json:"hotkey"`
	HasTransfer bool    `json:"has_transfer"`
	Paused      bool    `json:"paused"`
	Bytes       int     `json:"bytes"`
	Streams     int     `json:"streams"`
	FrameCount  int     `json:"frame_count"`
	Total       int     `json:"total"`
	FPS         int     `json:"fps"`
	CRCHex      string  `json:"crc_hex,omitempty"`
	CRC32       string  `json:"crc32,omitempty"`
	ETASeconds  float64 `json:"eta_seconds"`
	Error       string  `json:"error,omitempty"`
}

type controlRequest struct {
	Action string `json:"action"` // pause | resume | toggle
}

type sendResponse struct {
	Bytes      int     `json:"bytes"`
	FrameCount int     `json:"frame_count"`
	Frames     int     `json:"frames"`
	Streams    int     `json:"streams"`
	FPS        int     `json:"fps"`
	CRCHex     string  `json:"crc_hex,omitempty"`
	CRC32      string  `json:"crc32,omitempty"`
	ETASeconds float64 `json:"eta_seconds"`
	LoopSec    float64 `json:"loop_sec"`
	Popup      string  `json:"popup"`
	Error      string  `json:"error,omitempty"`
}

func (s *SenderServer) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	hk := htmlEscape(s.Hotkey)
	fmt.Fprintf(w, indexHTML,
		hk, hk,
		selectedAttr(s.Streams, 0),
		selectedAttr(s.Streams, 1),
		selectedAttr(s.Streams, 2),
		selectedAttr(s.Streams, 3),
		selectedAttr(s.Streams, 4),
	)
}

func selectedAttr(current, want int) string {
	if current == want {
		return " selected"
	}
	return ""
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
	standClass := ""
	if r.URL.Query().Get("stand") == "1" {
		standClass = "stand"
	}
	// body class, hint hotkey, JS hotkey string, stream count
	fmt.Fprintf(w, popupHTML, standClass, htmlEscape(s.Hotkey), s.Hotkey, streams)
}

func (s *SenderServer) handleLatest(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.latest == nil {
		writeJSON(w, http.StatusOK, encodeResponse{Error: "Clipboard is empty or no transfer yet — copy text, then press " + s.Hotkey})
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
		CRCHex:     s.latest.CRCHex,
		ETASeconds: s.latest.ETASeconds,
		Image:      s.latest.Image,
		Format:     "svg",
		Looping:    !s.latest.Static,
		Paused:     s.paused,
		Error:      s.latest.Error,
	}
	if r.URL.Query().Get("format") == "frames" && len(s.latest.Frames) > 0 {
		resp.Frames = s.latest.Frames
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *SenderServer) handleStatus(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.statusSnapshot())
}

func (s *SenderServer) statusSnapshot() statusResponse {
	s.mu.RLock()
	defer s.mu.RUnlock()
	resp := statusResponse{Hotkey: s.Hotkey, Paused: s.paused}
	if s.latest != nil {
		resp.HasTransfer = true
		resp.Bytes = s.latest.Bytes
		resp.Streams = s.latest.Streams
		resp.FrameCount = s.latest.FrameCount
		resp.Total = s.latest.FrameCount
		resp.FPS = s.latest.FPS
		resp.CRCHex = s.latest.CRCHex
		resp.CRC32 = s.latest.CRCHex
		resp.ETASeconds = s.latest.ETASeconds
		resp.Error = s.latest.Error
	}
	return resp
}

func (s *SenderServer) handleSend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	if err := s.showClipboard(); err != nil {
		msg := err.Error()
		if strings.Contains(strings.ToLower(msg), "empty") {
			msg = "Clipboard is empty — copy text first, then try again"
		}
		writeJSON(w, http.StatusBadRequest, sendResponse{Error: msg})
		return
	}
	tr := s.Latest()
	if tr == nil {
		writeJSON(w, http.StatusBadRequest, sendResponse{Error: "No transfer"})
		return
	}
	if tr.Error != "" {
		writeJSON(w, http.StatusBadRequest, sendResponse{Error: tr.Error})
		return
	}
	loopSec := 0.0
	if tr.FPS > 0 {
		loopSec = float64(tr.FrameCount) / float64(tr.FPS)
	}
	writeJSON(w, http.StatusOK, sendResponse{
		Bytes:      tr.Bytes,
		FrameCount: tr.FrameCount,
		Frames:     tr.FrameCount,
		Streams:    tr.Streams,
		FPS:        tr.FPS,
		CRCHex:     tr.CRCHex,
		CRC32:      tr.CRCHex,
		ETASeconds: tr.ETASeconds,
		LoopSec:    loopSec,
		Popup:      "/popup",
	})
}

func (s *SenderServer) handleSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "bad form"})
		return
	}
	msg := "Saved."
	if v := strings.TrimSpace(r.FormValue("streams")); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 || n > 4 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "streams must be 0–4"})
			return
		}
		s.mu.Lock()
		s.Streams = n
		s.mu.Unlock()
		msg = fmt.Sprintf("Streams set to %d (0 = auto).", n)
	}
	if hk := strings.TrimSpace(r.FormValue("hotkey")); hk != "" && !strings.EqualFold(hk, s.Hotkey) {
		msg += " Hotkey change requires restart with -hotkey " + hk + "."
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": msg})
}

func (s *SenderServer) handleControl(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	var req controlRequest
	_ = json.NewDecoder(r.Body).Decode(&req)
	s.mu.Lock()
	switch strings.ToLower(req.Action) {
	case "pause":
		s.paused = true
	case "resume":
		s.paused = false
	case "toggle":
		s.paused = !s.paused
	}
	paused := s.paused
	s.mu.Unlock()
	writeJSON(w, http.StatusOK, map[string]bool{"paused": paused})
}

func (s *SenderServer) SetPaused(v bool) {
	s.mu.Lock()
	s.paused = v
	s.mu.Unlock()
}

func (s *SenderServer) IsPaused() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.paused
}

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
		writeJSON(w, http.StatusBadRequest, encodeResponse{Error: "Invalid request — try pasting text again"})
		return
	}
	tr, err := s.buildTransfer(req.Text, req)
	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "empty") {
			msg = "Clipboard is empty — copy text first, then try again"
		}
		writeJSON(w, http.StatusBadRequest, encodeResponse{Error: msg})
		return
	}
	s.setLatest(tr)
	s.SetPaused(false)
	writeJSON(w, http.StatusOK, encodeResponse{
		FrameCount: tr.FrameCount,
		FPS:        tr.FPS,
		Bytes:      tr.Bytes,
		ChunkLen:   tr.ChunkLen,
		Redundancy: tr.Redundancy,
		Streams:    tr.Streams,
		Static:     tr.Static,
		CRCHex:     tr.CRCHex,
		ETASeconds: tr.ETASeconds,
		Image:      tr.Image,
		Frames:     tr.Frames,
		Format:     "svg",
		Looping:    !tr.Static,
		Paused:     false,
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

	crc := txqr.PayloadCRC([]byte(text))
	// Rough phone capture estimate: ~3 unique frames/sec single, ~5 with dual.
	capture := 3.0
	if streams >= 2 {
		capture = 5.0
	}
	needed := float64(len(chunks)) / used.Redundancy
	if needed < 1 {
		needed = 1
	}
	eta := needed / capture

	tr := &transfer{
		Text:       text,
		Bytes:      len(text),
		FrameCount: len(chunks),
		FPS:        used.FPS,
		ChunkLen:   used.ChunkLen,
		Redundancy: used.Redundancy,
		Streams:    streams,
		CRCHex:     fmt.Sprintf("%08x", crc),
		ETASeconds: eta,
		Created:    time.Now(),
		Chunks:     chunks,
		Static:     len(chunks) == 1,
	}

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

func (s *SenderServer) Latest() *transfer {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.latest
}

func (s *SenderServer) showClipboard() error {
	text, err := readClipboardText()
	if err != nil {
		tr := &transfer{Error: "Couldn't read clipboard — copy text, then try again", Created: time.Now()}
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
	s.SetPaused(false)
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
