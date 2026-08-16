package main

import (
	"fmt"

	"golang.design/x/clipboard"
)

func initClipboard() error {
	return clipboard.Init()
}

func readClipboardText() (string, error) {
	data := clipboard.Read(clipboard.FmtText)
	if data == nil {
		return "", fmt.Errorf("no text on clipboard")
	}
	return string(data), nil
}
