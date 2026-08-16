import Foundation
import UIKit

#if canImport(Txqr)
import Txqr
#endif

/// Owns the TXQR fountain decoder and exposes UI-friendly progress.
/// Vision multi-detect feeds every QR payload from one camera frame via DecodeBatch.
@MainActor
final class TransferSession: ObservableObject {
    @Published var progress: Int = 0
    @Published var statusText: String = "Ready — aim at the overlay"
    @Published var isComplete: Bool = false
    @Published var recoveredText: String = ""
    @Published var lastCodesInFrame: Int = 0
    @Published var preview: String = ""

    private var acceptedTotal: Int = 0

#if canImport(Txqr)
    private var decoder = TxqrNewDecoder()
#endif

    func start() {
        statusText = "Scanning for TXQR frames…"
    }

    func stop() {}

    /// Called on the main actor with all QR payloads seen in one camera frame.
    func ingest(codes: [String]) {
        lastCodesInFrame = codes.count
        guard !isComplete, !codes.isEmpty else { return }

        let joined = codes.joined(separator: "\n")

#if canImport(Txqr)
        let accepted = Int(decoder.decodeBatch(joined))
        if accepted > 0 {
            acceptedTotal += accepted
            progress = Int(decoder.progress())
            let unique = Int(decoder.uniqueFrames())
            statusText = "Unique frames \(unique) · last frame saw \(codes.count) QR"
            if decoder.isCompleted() {
                finish(text: decoder.data())
            }
        } else if codes.count > 0 {
            statusText = "Saw \(codes.count) QR — waiting for TXQR frames"
        }
#else
        // Framework not linked yet: keep UI usable while developing layout.
        statusText = "Txqr framework missing — run make ios-framework on a Mac"
        _ = joined
#endif
    }

    private func finish(text: String) {
        isComplete = true
        progress = 100
        recoveredText = text
        preview = text.count > 400 ? String(text.prefix(400)) + "…" : text
        statusText = "Complete · \(text.count) bytes"
        UINotificationFeedbackGenerator().notificationOccurred(.success)
    }

    func copyToClipboard() {
        UIPasteboard.general.string = recoveredText
        statusText = "Copied \(recoveredText.count) bytes"
    }

    func reset() {
#if canImport(Txqr)
        decoder.reset()
#endif
        progress = 0
        isComplete = false
        recoveredText = ""
        preview = ""
        acceptedTotal = 0
        lastCodesInFrame = 0
        statusText = "Scanning for TXQR frames…"
    }
}
