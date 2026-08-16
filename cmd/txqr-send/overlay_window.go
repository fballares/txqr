package main

import (
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

// openOverlayWindow opens a compact, movable QR overlay.
// On Windows it prefers Edge/Chrome app mode in the bottom-right corner
// and tries to keep the window always-on-top so a phone can sit on it.
func openOverlayWindow(url string, width, height int) error {
	if width <= 0 {
		width = 420
	}
	if height <= 0 {
		height = 520
	}
	switch runtime.GOOS {
	case "windows":
		if err := openWindowsOverlay(url, width, height); err == nil {
			return nil
		}
		return openBrowser(url)
	case "darwin":
		if err := exec.Command("open", "-na", "Google Chrome", "--args",
			"--app="+url,
			fmt.Sprintf("--window-size=%d,%d", width, height),
		).Start(); err == nil {
			return nil
		}
		return openBrowser(url)
	default:
		return openBrowser(url)
	}
}

func openWindowsOverlay(url string, width, height int) error {
	ps := `
$ErrorActionPreference = 'Stop'
$url = '` + psQuote(url) + `'
$w = ` + strconv.Itoa(width) + `
$h = ` + strconv.Itoa(height) + `
Add-Type -AssemblyName System.Windows.Forms
$screen = [System.Windows.Forms.Screen]::PrimaryScreen.WorkingArea
$x = [Math]::Max(0, $screen.Right - $w - 16)
$y = [Math]::Max(0, $screen.Bottom - $h - 16)

$candidates = @(
  "${env:ProgramFiles(x86)}\Microsoft\Edge\Application\msedge.exe",
  "$env:ProgramFiles\Microsoft\Edge\Application\msedge.exe",
  "$env:ProgramFiles\Google\Chrome\Application\chrome.exe",
  "${env:ProgramFiles(x86)}\Google\Chrome\Application\chrome.exe"
)
$browser = $candidates | Where-Object { Test-Path $_ } | Select-Object -First 1
if (-not $browser) { throw 'No Edge/Chrome found' }

$argList = @("--app=$url", "--window-size=$w,$h", "--window-position=$x,$y", "--new-window")
$proc = Start-Process -FilePath $browser -ArgumentList $argList -PassThru

Add-Type @"
using System;
using System.Runtime.InteropServices;
public static class TxqrWin {
  public static readonly IntPtr HWND_TOPMOST = new IntPtr(-1);
  public const uint SWP_SHOWWINDOW = 0x0040;
  [DllImport("user32.dll")] public static extern bool SetWindowPos(IntPtr hWnd, IntPtr after, int X, int Y, int cx, int cy, uint flags);
}
"@

$deadline = (Get-Date).AddSeconds(8)
$hwnd = [IntPtr]::Zero
while ((Get-Date) -lt $deadline) {
  Start-Sleep -Milliseconds 250
  try { $proc.Refresh() } catch {}
  if ($proc.MainWindowHandle -ne [IntPtr]::Zero) {
    $hwnd = $proc.MainWindowHandle
    break
  }
}
if ($hwnd -ne [IntPtr]::Zero) {
  [void][TxqrWin]::SetWindowPos($hwnd, [TxqrWin]::HWND_TOPMOST, $x, $y, $w, $h, [TxqrWin]::SWP_SHOWWINDOW)
}
`
	// Normalize newlines for -Command
	ps = strings.ReplaceAll(ps, "\r\n", "\n")
	cmd := exec.Command("powershell", "-NoProfile", "-ExecutionPolicy", "Bypass", "-Command", ps)
	return cmd.Start()
}

func psQuote(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}
