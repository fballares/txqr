//go:build windows

package main

import "golang.design/x/hotkey"

func modAlt() hotkey.Modifier { return hotkey.ModAlt }
func modWin() hotkey.Modifier { return hotkey.ModWin }
