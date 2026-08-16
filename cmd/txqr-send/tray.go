package main

import (
	_ "embed"
	"fmt"
	"log"
	"os"
	"sync/atomic"

	"fyne.io/systray"
)

//go:embed icon.ico
var trayIcon []byte

// trayApp wires the system tray to clipboard → QR actions.
type trayApp struct {
	hotkey   string
	baseURL  string
	status   *atomic.Value // string
	onShowQR func()
	onReplay func()
	onPause  func()
	onResume func()
	onStand  func()
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
	app.refreshTooltip("Ready")

	systray.SetOnTapped(func() {
		app.onShowQR()
	})

	mShow := systray.AddMenuItem("Show QR from clipboard", "Encode clipboard and open overlay")
	mShow.SetIcon(trayIcon)
	mReplay := systray.AddMenuItem("Show last transfer again", "Re-open overlay for the last payload")
	mStand := systray.AddMenuItem("Phone stand mode", "Bottom-right, minimal chrome, max contrast")
	systray.AddSeparator()
	mPause := systray.AddMenuItem("Pause animation", "Freeze the current QR frames")
	mResume := systray.AddMenuItem("Resume animation", "Continue the looping QR stream")
	systray.AddSeparator()
	mUI := systray.AddMenuItem("Open settings / coach…", "First-run guide and encoder settings")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit TXQR Send", "Stop the background sender")

	go func() {
		for {
			select {
			case <-mShow.ClickedCh:
				app.onShowQR()
			case <-mReplay.ClickedCh:
				if app.onReplay != nil {
					app.onReplay()
				}
			case <-mStand.ClickedCh:
				if app.onStand != nil {
					app.onStand()
				}
			case <-mPause.ClickedCh:
				if app.onPause != nil {
					app.onPause()
				}
				app.refreshTooltip("Paused")
			case <-mResume.ClickedCh:
				if app.onResume != nil {
					app.onResume()
				}
				app.refreshTooltip("Showing QR")
			case <-mUI.ClickedCh:
				app.onOpenUI()
			case <-mQuit.ClickedCh:
				systray.Quit()
				os.Exit(0)
			}
		}
	}()

	log.Printf("Tray ready — click or press %s · stand mode available", app.hotkey)
}

func (app *trayApp) refreshTooltip(state string) {
	if app.status != nil {
		app.status.Store(state)
	}
	systray.SetTooltip(fmt.Sprintf("TXQR Send · %s · hotkey %s", state, app.hotkey))
}
