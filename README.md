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

Use `cmd/txqr-send` as a background app:

1. Copy text on Windows  
2. Press **Ctrl+Shift+Q** (configurable)  
3. A popup shows a static or animated QR stream  
4. Your mobile TXQR reader decodes and copies the text  

```bash
go run ./cmd/txqr-send
# Windows GUI build:
go build -ldflags="-H windowsgui" -o txqr-send.exe ./cmd/txqr-send
```

Encoder defaults are auto-tuned for clipboard pastes (`txqr.ProfileForPayload`).

# Licence

MIT
