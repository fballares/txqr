// Command txqr-send runs in the background on Windows: copy text, press a
// global hotkey or click the tray icon, and a popup shows the TXQR stream.
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/divan/txqr"
)

func main() {
	defaults := txqr.DefaultClipboardProfile()

	addr := flag.String("addr", "127.0.0.1:1988", "Local address for the QR popup UI")
	hotkeySpec := flag.String("hotkey", "Ctrl+Shift+Q", "Global hotkey to show QR for clipboard text")
	chunk := flag.Int("split", defaults.ChunkLen, "Override chunk size (0 = auto from payload)")
	fps := flag.Int("fps", defaults.FPS, "Override animation FPS")
	size := flag.Int("size", defaults.QRSize, "QR image size in pixels")
	redundancy := flag.Float64("redundancy", defaults.Redundancy, "Fountain-code redundancy factor")
	background := flag.Bool("background", true, "Stay resident with tray icon and hotkey")
	tray := flag.Bool("tray", true, "Show a system tray / notification-area icon")
	text := flag.String("text", "", "Encode this text once and show popup (then exit unless -background)")
	noBrowser := flag.Bool("n", false, "Do not open a browser automatically")
	flag.Parse()

	profile := txqr.ClipboardProfile{
		ChunkLen:   *chunk,
		FPS:        *fps,
		QRSize:     *size,
		Redundancy: *redundancy,
	}

	if err := initClipboard(); err != nil {
		log.Printf("clipboard init: %v (clipboard reads may fail)", err)
	}

	srv := &SenderServer{
		Default: profile,
		Hotkey:  *hotkeySpec,
	}

	ln, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("listen %s: %v", *addr, err)
	}
	baseURL := fmt.Sprintf("http://%s", ln.Addr().String())

	mux := http.NewServeMux()
	mux.HandleFunc("/", srv.handleIndex)
	mux.HandleFunc("/popup", srv.handlePopup)
	mux.HandleFunc("/api/encode", srv.handleEncode)
	mux.HandleFunc("/api/latest", srv.handleLatest)

	go func() {
		if err := http.Serve(ln, mux); err != nil {
			log.Fatal(err)
		}
	}()

	initial := strings.TrimSpace(*text)
	if initial == "" && !isInteractive() {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			log.Fatalf("read stdin: %v", err)
		}
		initial = string(data)
	}

	openPopup := func() {
		url := fmt.Sprintf("%s/popup?t=%d", baseURL, time.Now().UnixNano())
		if err := openBrowser(url); err != nil {
			log.Printf("open popup: %v (open %s manually)", err, url)
		}
	}

	triggerFromClipboard := func() {
		if err := srv.showClipboard(); err != nil {
			log.Printf("clipboard transfer: %v", err)
		} else {
			srv.mu.RLock()
			tr := srv.latest
			srv.mu.RUnlock()
			if tr != nil && tr.Error == "" {
				log.Printf("showing QR: %d bytes, %d frames @ %d fps", tr.Bytes, tr.FrameCount, tr.FPS)
			}
		}
		if !*noBrowser {
			openPopup()
		}
	}

	if initial != "" {
		tr, err := srv.buildTransfer(initial, encodeRequest{Auto: true})
		if err != nil {
			log.Fatalf("encode: %v", err)
		}
		srv.setLatest(tr)
		if !*noBrowser {
			time.Sleep(200 * time.Millisecond)
			openPopup()
		}
		if !*background {
			log.Printf("Transfer ready at %s/popup — Ctrl+C to exit", baseURL)
			select {}
		}
	}

	log.Printf("TXQR sender running at %s", baseURL)
	log.Printf("Workflow: copy text → click tray icon or press %s → scan with phone", *hotkeySpec)
	log.Printf("Defaults: chunk=%d fps=%d size=%d redundancy=%.2f (auto-tuned per paste size)",
		profile.ChunkLen, profile.FPS, profile.QRSize, profile.Redundancy)

	if *background {
		go runHotkeyLoop(*hotkeySpec, triggerFromClipboard)

		if *tray {
			runTray(&trayApp{
				hotkey:  *hotkeySpec,
				baseURL: baseURL,
				onShowQR: func() {
					triggerFromClipboard()
				},
				onOpenUI: func() {
					_ = openBrowser(baseURL + "/")
				},
			})
			return
		}

		// No tray: hotkey-only resident mode.
		select {}
	}

	if !*noBrowser && initial == "" {
		time.Sleep(200 * time.Millisecond)
		_ = openBrowser(baseURL + "/")
	}
	select {}
}

func isInteractive() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return true
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}

func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}
