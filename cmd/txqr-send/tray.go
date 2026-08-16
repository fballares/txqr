package main

import (
	_ "embed"
	"fmt"
	"log"
	"os"

	"fyne.io/systray"
)

//go:embed icon.ico
var trayIcon []byte

// trayApp wires the system tray to clipboard → QR actions.
type trayApp struct {
	hotkey   string
	baseURL  string
	onShowQR func()
	onOpenUI func()
}

func runTray(app *trayApp) {
	systray.Run(func() { app.onReady() }, func() {
		log.Printf("TXQR sender exiting")
	})
}

func (app *trayApp) onReady() {
	systray.SetIcon(trayIcon)
	systray.SetTitle("TXQR")
	systray.SetTooltip(fmt.Sprintf("TXQR Send — copy text, click tray or press %s (overlay loops until done)", app.hotkey))

	// Left-click tray icon = generate QR from clipboard
	systray.SetOnTapped(func() {
		app.onShowQR()
	})

	mShow := systray.AddMenuItem("Show QR from clipboard", "Encode clipboard and open movable overlay")
	mShow.SetIcon(trayIcon)
	mUI := systray.AddMenuItem("Open settings…", "Open the local settings page")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit TXQR Send", "Stop the background sender")

	go func() {
		for {
			select {
			case <-mShow.ClickedCh:
				app.onShowQR()
			case <-mUI.ClickedCh:
				app.onOpenUI()
			case <-mQuit.ClickedCh:
				systray.Quit()
				// Ensure process exits even if other goroutines hang.
				os.Exit(0)
			}
		}
	}()

	log.Printf("Tray icon ready — click it or press %s to show QR", app.hotkey)
}
