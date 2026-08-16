#!/usr/bin/env bash
# Generate TXQRReader.xcodeproj (requires XcodeGen on macOS).
set -euo pipefail
cd "$(dirname "$0")"
if ! command -v xcodegen >/dev/null 2>&1; then
  echo "Install XcodeGen: brew install xcodegen" >&2
  exit 1
fi
xcodegen generate
echo "Open with: open TXQRReader.xcodeproj"
echo "Also run from repo root: make ios-framework"
