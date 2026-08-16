// Command txqr-send shows selected/pasted text as a looping animated QR
// stream so a mobile TXQR reader can capture and copy it.
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
)

func main() {
	addr := flag.String("addr", "127.0.0.1:1988", "Local address to serve the sender UI")
	chunk := flag.Int("split", 120, "Payload bytes per QR frame")
	fps := flag.Int("fps", 8, "Animation frames per second")
	size := flag.Int("size", 480, "QR image size in pixels")
	text := flag.String("text", "", "Text to send immediately (skips paste step)")
	noBrowser := flag.Bool("n", false, "Do not open a browser automatically")
	flag.Parse()

	initial := strings.TrimSpace(*text)
	if initial == "" && !isInteractive() {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			log.Fatalf("read stdin: %v", err)
		}
		initial = string(data)
	}

	srv := &SenderServer{
		ChunkLen: *chunk,
		FPS:      *fps,
		QRSize:   *size,
		Initial:  initial,
	}

	ln, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("listen %s: %v", *addr, err)
	}

	url := fmt.Sprintf("http://%s/", ln.Addr().String())
	log.Printf("TXQR sender ready at %s", url)
	log.Printf("Paste or pass text, then point your phone camera at the animated QR stream.")

	if !*noBrowser {
		go func() {
			time.Sleep(300 * time.Millisecond)
			if err := openBrowser(url); err != nil {
				log.Printf("open browser: %v (open %s manually)", err, url)
			}
		}()
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", srv.handleIndex)
	mux.HandleFunc("/api/encode", srv.handleEncode)

	if err := http.Serve(ln, mux); err != nil {
		log.Fatal(err)
	}
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

