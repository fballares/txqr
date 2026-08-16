package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestIndexAndPopupRender(t *testing.T) {
	srv := &SenderServer{Hotkey: "Ctrl+Shift+Q", Streams: 0}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	srv.handleIndex(rec, req)
	if rec.Code != 200 {
		t.Fatalf("index status %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Ctrl+Shift+Q") {
		t.Fatalf("index missing hotkey")
	}
	if strings.Contains(body, "%!") {
		t.Fatalf("index format error")
	}

	tr, err := srv.buildTransfer(strings.Repeat("hello world ", 200), encodeRequest{Auto: true, Streams: 2})
	if err != nil {
		t.Fatal(err)
	}
	srv.setLatest(tr)

	req = httptest.NewRequest(http.MethodGet, "/popup?stand=1", nil)
	rec = httptest.NewRecorder()
	srv.handlePopup(rec, req)
	if rec.Code != 200 {
		t.Fatalf("popup status %d", rec.Code)
	}
	body = rec.Body.String()
	if !strings.Contains(body, `class="stand"`) {
		t.Fatalf("expected stand class")
	}
	if !strings.Contains(body, "Ctrl+Shift+Q") {
		t.Fatalf("popup missing hotkey")
	}
	if strings.Contains(body, "%!") {
		t.Fatalf("popup format error")
	}

	req = httptest.NewRequest(http.MethodGet, "/api/status", nil)
	rec = httptest.NewRecorder()
	srv.handleStatus(rec, req)
	if rec.Code != 200 {
		t.Fatalf("status %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"total"`) {
		t.Fatalf("status missing total: %s", rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodPost, "/api/control", strings.NewReader(`{"action":"pause"}`))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	srv.handleControl(rec, req)
	if rec.Code != 200 {
		t.Fatalf("control %d", rec.Code)
	}
	if !srv.IsPaused() {
		t.Fatal("expected paused")
	}

	_ = fmt.Sprintf("ok frames=%d streams=%d", tr.FrameCount, tr.Streams)
}
