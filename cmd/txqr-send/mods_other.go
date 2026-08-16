//go:build !windows && !darwin && !linux

package main

import "golang.design/x/hotkey"

func modAlt() hotkey.Modifier { return 0 }
func modWin() hotkey.Modifier { return 0 }
