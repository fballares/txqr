//go:build darwin

package main

import "golang.design/x/hotkey"

func modAlt() hotkey.Modifier { return hotkey.ModOption }
func modWin() hotkey.Modifier { return hotkey.ModCmd }
