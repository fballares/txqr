package main

import (
	"fmt"
	"log"
	"strings"

	"golang.design/x/hotkey"
)

// parseHotkey parses strings like "Ctrl+Shift+Q" into modifiers + key.
func parseHotkey(spec string) ([]hotkey.Modifier, hotkey.Key, error) {
	parts := strings.Split(spec, "+")
	if len(parts) < 2 {
		return nil, 0, errHotkey("need modifiers and a key, e.g. Ctrl+Shift+Q")
	}
	var mods []hotkey.Modifier
	var key hotkey.Key
	for i, p := range parts {
		p = strings.TrimSpace(p)
		if i == len(parts)-1 {
			k, ok := keyFromName(p)
			if !ok {
				return nil, 0, errHotkey("unknown key %q", p)
			}
			key = k
			continue
		}
		switch strings.ToLower(p) {
		case "ctrl", "control":
			mods = append(mods, hotkey.ModCtrl)
		case "shift":
			mods = append(mods, hotkey.ModShift)
		case "alt", "option":
			mods = append(mods, modAlt())
		case "win", "super", "cmd", "meta":
			mods = append(mods, modWin())
		default:
			return nil, 0, errHotkey("unknown modifier %q", p)
		}
	}
	if len(mods) == 0 {
		return nil, 0, errHotkey("at least one modifier required")
	}
	return mods, key, nil
}

func keyFromName(name string) (hotkey.Key, bool) {
	switch strings.ToUpper(strings.TrimSpace(name)) {
	case "A":
		return hotkey.KeyA, true
	case "B":
		return hotkey.KeyB, true
	case "C":
		return hotkey.KeyC, true
	case "D":
		return hotkey.KeyD, true
	case "E":
		return hotkey.KeyE, true
	case "F":
		return hotkey.KeyF, true
	case "G":
		return hotkey.KeyG, true
	case "H":
		return hotkey.KeyH, true
	case "I":
		return hotkey.KeyI, true
	case "J":
		return hotkey.KeyJ, true
	case "K":
		return hotkey.KeyK, true
	case "L":
		return hotkey.KeyL, true
	case "M":
		return hotkey.KeyM, true
	case "N":
		return hotkey.KeyN, true
	case "O":
		return hotkey.KeyO, true
	case "P":
		return hotkey.KeyP, true
	case "Q":
		return hotkey.KeyQ, true
	case "R":
		return hotkey.KeyR, true
	case "S":
		return hotkey.KeyS, true
	case "T":
		return hotkey.KeyT, true
	case "U":
		return hotkey.KeyU, true
	case "V":
		return hotkey.KeyV, true
	case "W":
		return hotkey.KeyW, true
	case "X":
		return hotkey.KeyX, true
	case "Y":
		return hotkey.KeyY, true
	case "Z":
		return hotkey.KeyZ, true
	case "0":
		return hotkey.Key0, true
	case "1":
		return hotkey.Key1, true
	case "2":
		return hotkey.Key2, true
	case "3":
		return hotkey.Key3, true
	case "4":
		return hotkey.Key4, true
	case "5":
		return hotkey.Key5, true
	case "6":
		return hotkey.Key6, true
	case "7":
		return hotkey.Key7, true
	case "8":
		return hotkey.Key8, true
	case "9":
		return hotkey.Key9, true
	case "SPACE":
		return hotkey.KeySpace, true
	case "ENTER", "RETURN":
		return hotkey.KeyReturn, true
	case "TAB":
		return hotkey.KeyTab, true
	case "ESC", "ESCAPE":
		return hotkey.KeyEscape, true
	}
	return 0, false
}

type hotkeyError string

func (e hotkeyError) Error() string { return string(e) }

func errHotkey(format string, args ...interface{}) error {
	return hotkeyError(fmt.Sprintf(format, args...))
}

// runHotkeyLoop registers the global hotkey and invokes onTrigger on each press.
// Blocks until the process exits.
func runHotkeyLoop(spec string, onTrigger func()) {
	mods, key, err := parseHotkey(spec)
	if err != nil {
		log.Printf("hotkey: %v — background hotkey disabled", err)
		return
	}
	hk := hotkey.New(mods, key)
	if err := hk.Register(); err != nil {
		log.Printf("hotkey register %s failed: %v — use the UI or /popup manually", spec, err)
		return
	}
	log.Printf("Hotkey ready: %s (copy text, then press to show QR)", spec)
	for range hk.Keydown() {
		onTrigger()
	}
}
