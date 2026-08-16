# txqr-send

Background Windows sender for clipboard → animated QR → phone reader.

## Workflow

1. Start once and leave it running:
   ```bash
   go run ./cmd/txqr-send
   ```
2. Select text anywhere and copy it (`Ctrl+C`).
3. Press **`Ctrl+Shift+Q`** (configurable).
4. A popup window shows a static QR (short text) or a looping animated TXQR stream.
5. Point your mobile TXQR reader at the window; when decode completes, copy the text on the phone.

## Optimal defaults (clipboard use case)

| Setting | Default | Why |
|---------|---------|-----|
| ECC | Medium | Survives focus/angle noise without oversized modules |
| Chunk | **auto** (~110–150 bytes) | Mid QR versions decode reliably from a monitor |
| FPS | **5–6** | Matches typical phone decode speed; faster drops frames |
| Redundancy | **2.0–2.5** | Fountain codes recover missed frames |
| QR size | **560–640px** | Fills a popup for easy camera lock |

Payload size auto-tunes chunk/FPS/redundancy via `txqr.ProfileForPayload`.

## Flags

```bash
txqr-send -hotkey Ctrl+Shift+Q
txqr-send -background=false -text "hello"
txqr-send -split 0 -fps 6 -size 560
```

| Flag | Default | Meaning |
|------|---------|---------|
| `-hotkey` | `Ctrl+Shift+Q` | Global shortcut |
| `-background` | `true` | Stay resident for hotkey |
| `-addr` | `127.0.0.1:1988` | Local UI |
| `-split` | `140` | Chunk override (`0` in API = auto) |
| `-fps` | `6` | Animation FPS |
| `-size` | `560` | QR pixel size |
| `-redundancy` | `2.0` | Fountain redundancy |
| `-n` | false | Don’t auto-open browser |

### Windows build (no console window)

```bash
go build -ldflags="-H windowsgui" -o txqr-send.exe ./cmd/txqr-send
```

Put `txqr-send.exe` in your Startup folder to run at login.

## Mobile reader

Use `github.com/divan/txqr/mobile`:

1. `Decode(frame)` for each scanned QR
2. When `IsCompleted()`, `Data()` → clipboard
3. `Reset()` before the next transfer
