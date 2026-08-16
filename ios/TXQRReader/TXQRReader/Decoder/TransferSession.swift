import Foundation
import UIKit

#if canImport(Txqr)
import Txqr
#endif

/// Owns the TXQR fountain decoder and exposes UI-friendly progress.
/// Dynamically accepts 1 or many QR payloads per camera frame — no fixed stream count.
@MainActor
final class TransferSession: ObservableObject {
    enum DetectedLayout: String {
        case idle = "Waiting"
        case single = "Single QR"
        case dual = "Dual QR"
        case multi = "Multi QR"
    }

    @Published var progress: Int = 0
    @Published var statusText: String = "Ready — aim at the overlay"
    @Published var isComplete: Bool = false
    @Published var recoveredText: String = ""
    @Published var lastCodesInFrame: Int = 0
    @Published var peakConcurrent: Int = 0
    @Published var detectedLayout: DetectedLayout = .idle
    @Published var preview: String = ""

    private var acceptedTotal: Int = 0

#if canImport(Txqr)
    private var decoder = TxqrNewDecoder()
#endif

    func start() {
        statusText = "Scanning — auto-detects 1 or 2 QR codes"
    }

    func stop() {}

    /// Called with every QR payload Vision found in one camera frame.
    func ingest(codes: [String]) {
        lastCodesInFrame = codes.count
        if codes.count > peakConcurrent {
            peakConcurrent = codes.count
        }
        updateLayout(codes.count)

        guard !isComplete, !codes.isEmpty else { return }

        let joined = codes.joined(separator: "\n")

#if canImport(Txqr)
        let accepted = Int(decoder.decodeBatch(joined))
        if accepted > 0 {
            acceptedTotal += accepted
            progress = Int(decoder.progress())
            let unique = Int(decoder.uniqueFrames())
            let layout = detectedLayout.rawValue
            statusText = "\(layout) · \(unique) unique frames · +\(accepted) this tick"
            if decoder.isCompleted() {
                finish(text: decoder.data())
            }
        } else if codes.count > 0 {
            statusText = "Saw \(codes.count) QR — waiting for TXQR frames"
        }
#else
        statusText = "Txqr framework missing — run make ios-framework on a Mac"
        _ = joined
#endif
    }

    private func updateLayout(_ count: Int) {
        switch count {
        case 0:
            break
        case 1:
            // Don't downgrade from dual if we already locked onto two —
            // brief misses of one code are common while panning.
            if detectedLayout != .dual && detectedLayout != .multi {
                detectedLayout = .single
            }
        case 2:
            detectedLayout = .dual
        default:
            detectedLayout = .multi
        }
    }

    private func finish(text: String) {
        isComplete = true
        progress = 100
        recoveredText = text
        preview = text.count > 400 ? String(text.prefix(400)) + "…" : text
        let mode = peakConcurrent >= 2 ? "dual/multi peak \(peakConcurrent)" : "single"
        statusText = "Complete · \(text.count) bytes · \(mode)"
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
        peakConcurrent = 0
        detectedLayout = .idle
        statusText = "Scanning — auto-detects 1 or 2 QR codes"
    }
}
