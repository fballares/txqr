# txqr-send

Windows-friendly sender that turns selected/pasted text into a looping animated QR stream for a mobile TXQR reader.

## Usage

```bash
go run ./cmd/txqr-send
```

Then paste text in the browser UI and click **Show QR stream**. Point your phone at the animation; when decoding finishes, copy the recovered text in the mobile app.

### Options

```bash
txqr-send -text "hello from Windows"
txqr-send -split 120 -fps 8 -size 480
echo "pipe text" | txqr-send -n
```

| Flag | Default | Meaning |
|------|---------|---------|
| `-addr` | `127.0.0.1:1988` | Local UI address |
| `-split` | `120` | Bytes of payload per QR frame |
| `-fps` | `8` | Animation speed |
| `-size` | `480` | QR pixel size |
| `-text` | | Send this text immediately |
| `-n` | false | Do not auto-open a browser |

## Mobile reader

Use the Gomobile package `github.com/divan/txqr/mobile`:

1. Continuously decode scanned QR payloads with `Decoder.Decode(frame)`
2. When `IsCompleted()` is true, call `Data()` and copy to the clipboard
3. Call `Reset()` before the next transfer

Frame format produced by this sender:

```text
blockCode/chunkLen/total|payload
```
