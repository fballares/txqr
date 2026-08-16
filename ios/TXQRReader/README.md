# TXQRReader (iOS)

SwiftUI app for **iPhone** that reads **multiple TXQR QR codes concurrently** from the camera (Vision multi-detect), feeds them into the Gomobile `txqr` decoder, then copies or shares the recovered text/file bytes.

## Architecture

```
Windows txqr-send overlay          iPhone TXQRReader
┌─────────────────────┐            ┌──────────────────────────┐
│ QR A  QR B (loop)   │  light →   │ AVCaptureSession         │
│ same fountain set   │            │ VNDetectBarcodesRequest  │
└─────────────────────┘            │  (all .qr observations)  │
                                   │ DecodeBatch(joined)      │
                                   │ → clipboard / Files      │
                                   └──────────────────────────┘
```

Fountain frames are unchanged (`blockCode/chunkLen/total|payload`). Multiple on-screen QRs simply deliver more blocks per camera frame.

## Requirements

- macOS with Xcode 15+
- [Go](https://go.dev) 1.22+
- [gomobile](https://pkg.go.dev/golang.org/x/mobile/cmd/gomobile)
- Physical iPhone recommended (camera); Simulator cannot exercise real multi-QR well

## Build the Go framework

From the repo root on a Mac:

```bash
go install golang.org/x/mobile/cmd/gomobile@latest
gomobile init
make ios-framework
```

This writes `ios/TXQRReader/Frameworks/Txqr.xcframework` (or `txqr.framework` depending on gomobile version).

## Open & run the app

```bash
cd ios/TXQRReader
open TXQRReader.xcodeproj
```

1. Select your **Team** under Signing
2. Run on your **iPhone 17 Pro**
3. Allow Camera access
4. Point at the Windows multi-QR overlay; progress fills until Complete
5. Tap **Copy** (or **Share / Save**)

If you use [XcodeGen](https://github.com/yonaskolb/XcodeGen):

```bash
brew install xcodegen
cd ios/TXQRReader && xcodegen generate && open TXQRReader.xcodeproj
```

## Usage with Windows sender

```bash
# on Windows
txqr-send.exe -streams 2
```

Copy text → tray click / hotkey → overlay shows **two looping QRs** → iPhone reads both.

## Notes

- Non-TXQR codes in the scene are ignored by `DecodeBatch`
- You can start mid-loop; no need to catch frame 1
- 2 concurrent QRs is the default sweet spot; 3–4 are supported but optically harder
