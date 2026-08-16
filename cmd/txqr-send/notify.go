package main

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// notifyUser shows a brief OS notification (Windows toast when possible).
func notifyUser(title, body string) {
	switch runtime.GOOS {
	case "windows":
		_ = windowsToast(title, body)
	case "darwin":
		script := fmt.Sprintf(`display notification %q with title %q`, body, title)
		_ = exec.Command("osascript", "-e", script).Start()
	default:
		// Best-effort on Linux desktops.
		_ = exec.Command("notify-send", title, body).Start()
	}
}

func windowsToast(title, body string) error {
	// PowerShell WinRT toast — works on modern Windows without extra deps.
	ps := `
[Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType = WindowsRuntime] | Out-Null
[Windows.Data.Xml.Dom.XmlDocument, Windows.Data.Xml.Dom.XmlDocument, ContentType = WindowsRuntime] | Out-Null
$template = @"
<toast>
  <visual>
    <binding template="ToastGeneric">
      <text>__TITLE__</text>
      <text>__BODY__</text>
    </binding>
  </visual>
</toast>
"@
$xml = New-Object Windows.Data.Xml.Dom.XmlDocument
$xml.LoadXml($template)
$toast = [Windows.UI.Notifications.ToastNotification]::new($xml)
[Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier("TXQR Send").Show($toast)
`
	ps = strings.ReplaceAll(ps, "__TITLE__", xmlEscape(title))
	ps = strings.ReplaceAll(ps, "__BODY__", xmlEscape(body))
	return exec.Command("powershell", "-NoProfile", "-ExecutionPolicy", "Bypass", "-Command", ps).Start()
}

func xmlEscape(s string) string {
	r := strings.NewReplacer(
		`&`, "&amp;",
		`<`, "&lt;",
		`>`, "&gt;",
		`"`, "&quot;",
		`'`, "&apos;",
	)
	return r.Replace(s)
}
