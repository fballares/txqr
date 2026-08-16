//go:build linux

package main

import "golang.design/x/hotkey"

// On X11, Mod1 is typically Alt and Mod4 is Super/Win.
func modAlt() hotkey.Modifier { return hotkey.Mod1 }
func modWin() hotkey.Modifier { return hotkey.Mod4 }
