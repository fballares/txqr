package main

import (
	"encoding/json"
	"fmt"
	"net/http"
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

	// latest transfer prepared for /popup
	latest *transfer
}

type transfer struct {
	Text       string
	Bytes      int
	FrameCount int
	FPS        int
	ChunkLen   int
	Redundancy float64
	Created    time.Time
	Image      string // data URL (gif or png)
	Static     bool
	Error      string
}

type encodeRequest struct {
	Text       string  `json:"text"`
	ChunkLen   int     `json:"chunk_len"`
	FPS        int     `json:"fps"`
	QRSize     int     `json:"qr_size"`
	Redundancy float64 `json:"redundancy"`
	Auto       bool    `json:"auto"`
}

type encodeResponse struct {
	FrameCount int     `json:"frame_count"`
	FPS        int     `json:"fps"`
	Bytes      int     `json:"bytes"`
	ChunkLen   int     `json:"chunk_len"`
	Redundancy float64 `json:"redundancy"`
	Static     bool    `json:"static"`
	Image      string  `json:"image"`
	Error      string  `json:"error,omitempty"`
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
	fmt.Fprintf(w, popupHTML, htmlEscape(s.Hotkey))
}

func (s *SenderServer) handleLatest(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.latest == nil {
		writeJSON(w, http.StatusOK, encodeResponse{Error: "no transfer yet — copy text and press " + s.Hotkey})
		return
	}
	writeJSON(w, http.StatusOK, encodeResponse{
		FrameCount: s.latest.FrameCount,
		FPS:        s.latest.FPS,
		Bytes:      s.latest.Bytes,
		ChunkLen:   s.latest.ChunkLen,
		Redundancy: s.latest.Redundancy,
		Static:     s.latest.Static,
		Image:      s.latest.Image,
		Error:      s.latest.Error,
	})
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
		Static:     tr.Static,
		Image:      tr.Image,
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

	chunks, used, err := encodeTransfer(text, p)
	if err != nil {
		return &transfer{Error: err.Error(), Created: time.Now()}, err
	}

	tr := &transfer{
		Text:       text,
		Bytes:      len(text),
		FrameCount: len(chunks),
		FPS:        used.FPS,
		ChunkLen:   used.ChunkLen,
		Redundancy: used.Redundancy,
		Created:    time.Now(),
		Static:     len(chunks) == 1,
	}

	if tr.Static {
		pngBytes, err := renderPNG(chunks[0], used.QRSize)
		if err != nil {
			return nil, err
		}
		tr.Image = dataURL("image/png", pngBytes)
	} else {
		gifBytes, err := renderGIF(chunks, used.QRSize, used.FPS)
		if err != nil {
			return nil, err
		}
		tr.Image = dataURL("image/gif", gifBytes)
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
