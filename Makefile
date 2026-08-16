gomobile:
	gomobile bind -target=ios -o txqr.framework github.com/divan/txqr/mobile

# Build an XCFramework for the iOS TXQRReader app (run on macOS).
ios-framework:
	mkdir -p ios/TXQRReader/Frameworks
	gomobile bind -target=ios -o ios/TXQRReader/Frameworks/Txqr.xcframework github.com/divan/txqr/mobile

.PHONY: gomobile ios-framework
