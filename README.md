# TXQR

[![GoDoc](https://godoc.org/github.com/divan/txqr?status.svg)](https://godoc.org/github.com/divan/txqr)

TXQR (Transfer via QR) is a protocol and set of tools and libs to transfer data via animated QR codes. It uses [fountain codes](https://en.wikipedia.org/wiki/Fountain_code) for error correction.

See related blog posts for more details:
 - [Animated QR data transfer with Gomobile and Gopherjs](https://divan.github.io/posts/animatedqr/)
 - [Fountain codes and animated QR](https://divan.github.io/posts/fountaincodes/)

# Demo

![Demo](./docs/demo.gif)

Reader iOS app in the demo (uses this lib via Gomobile): [https://github.com/divan/txqr-reader](https://github.com/divan/txqr-reader)

## Automated tester app
Also see `cmd/txqr-tester` app for automated testing of different encoder parameters.

## Windows text → phone clipboard

Use `cmd/txqr-send` on Windows and `ios/TXQRReader` on iPhone:

1. Run `txqr-send` (tray icon, multi-QR overlay)  
2. Copy text → click tray or **Ctrl+Shift+Q**  
3. Overlay loops **1–2 QR codes** (drag to your phone stand)  
4. iPhone TXQRReader multi-detects all visible QRs and reconstructs the payload  

```bash
go run ./cmd/txqr-send          # auto 1 vs 2 QR by benefit
go run ./cmd/txqr-send -streams 2   # force dual
# Windows GUI build:
go build -ldflags="-H windowsgui" -o txqr-send.exe ./cmd/txqr-send

# On a Mac, build the Go framework for the iOS app:
make ios-framework
```

See `cmd/txqr-send/AIRGAP.md` and `ios/TXQRReader/README.md`.

# Licence

MIT
