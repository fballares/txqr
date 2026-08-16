import Foundation
import UIKit
import AudioToolbox

#if canImport(Txqr)
import Txqr
#endif

/// Owns the TXQR fountain decoder and exposes UI-friendly progress.
/// Dual mode: Windows shows LEFT + RIGHT slots; we sort Vision hits the same way
/// and ingest both payloads into one fountain decoder each camera tick.
@MainActor
final class TransferSession: ObservableObject {
    enum DetectedLayout: String {
        case idle = "Waiting"
        case single = "Single QR"
        case dual = "LEFT + RIGHT"
        case multi = "Multi QR"
    }

    @Published var progress: Int = 0
    @Published var statusText: String = "Ready — aim at the overlay"
    @Published var isComplete: Bool = false
    @Published var recoveredText: String = ""
    @Published var lastCodesInFrame: Int = 0
    @Published var peakConcurrent: Int = 0
    @Published var detectedLayout: DetectedLayout = .idle
    @Published var leftActive: Bool = false
    @Published var rightActive: Bool = false
    @Published var preview: String = ""
    @Published var integrityRetry: Bool = false

    private var acceptedTotal: Int = 0
    private var lastProgressHaptic = 0
    private let softHaptic = UIImpactFeedbackGenerator(style: .light)
    private let notifyHaptic = UINotificationFeedbackGenerator()

#if canImport(Txqr)
    private var decoder = TxqrNewDecoder()
#endif

    func start() {
        softHaptic.prepare()
        notifyHaptic.prepare()
        statusText = "Scanning — LEFT/RIGHT dual or single auto"
        TransferLiveActivityStub.start(status: statusText)
    }

    func stop() {
        TransferLiveActivityStub.end()
    }

    /// Positioned QR payloads already sorted LEFT → RIGHT by the scanner.
    func ingest(positioned: [PositionedQR]) {
        lastCodesInFrame = positioned.count
        if positioned.count > peakConcurrent {
            peakConcurrent = positioned.count
        }

        if positioned.count >= 2 {
            leftActive = true
            rightActive = true
            detectedLayout = positioned.count == 2 ? .dual : .multi
        } else if positioned.count == 1 {
            let x = positioned[0].midX
            leftActive = x < 0.55
            rightActive = x >= 0.45
            if detectedLayout != .dual && detectedLayout != .multi {
                detectedLayout = .single
            }
        } else {
            leftActive = false
            rightActive = false
        }

        guard !isComplete, !positioned.isEmpty else { return }

        let codes = positioned.map(\.payload)
        let joined = codes.joined(separator: "\n")

#if canImport(Txqr)
        let accepted = Int(decoder.decodeBatch(joined))
        if accepted > 0 {
            acceptedTotal += accepted
            let next = Int(decoder.progress())
            if next > progress {
                progress = next
                if next / 20 > lastProgressHaptic / 20 {
                    lastProgressHaptic = next
                    softHaptic.impactOccurred()
                }
            }
            let unique = Int(decoder.uniqueFrames())
            integrityRetry = false
            if positioned.count >= 2 {
                statusText = "LEFT+RIGHT locked · \(unique) unique · +\(accepted)"
            } else {
                let side = positioned[0].midX < 0.5 ? "LEFT" : "RIGHT"
                statusText = "\(side) QR · \(unique) unique · +\(accepted)"
            }
            if decoder.isCompleted() {
                let verified = decoder.isVerified()
                finish(text: decoder.data(), verified: verified, crc: decoder.expectedCRCHex())
            } else if !decoder.integrityError().isEmpty {
                integrityRetry = true
                statusText = "Integrity retry — \(decoder.integrityError())"
                notifyHaptic.notificationOccurred(.warning)
            }
        } else if !codes.isEmpty {
            statusText = "Saw \(codes.count) QR — waiting for TXQR frames"
        }
#else
        statusText = "Txqr framework missing — run make ios-framework on a Mac"
        _ = joined
#endif
    }

    private func finish(text: String, verified: Bool, crc: String) {
        isComplete = true
        progress = 100
        recoveredText = text
        preview = text.count > 400 ? String(text.prefix(400)) + "…" : text
        let mode = peakConcurrent >= 2 ? "LEFT+RIGHT peak \(peakConcurrent)" : "single"
        if verified {
            let crcPart = crc.isEmpty ? "verified" : "CRC \(crc) verified"
            statusText = "Complete · \(text.count) bytes · \(crcPart) · \(mode)"
        } else {
            statusText = "Complete · \(text.count) bytes · \(mode)"
        }
        notifyHaptic.notificationOccurred(.success)
        AudioServicesPlaySystemSound(1057)
        TransferLiveActivityStub.update(progress: 100, status: statusText)
    }

    func copyToClipboard() {
        UIPasteboard.general.string = recoveredText
        statusText = "Copied \(recoveredText.count) bytes"
    }

    func playCopyHaptic() {
        notifyHaptic.notificationOccurred(.success)
    }

    /// Writes recovered text to a temp .txt for Share / Files.
    func writeTempFile() -> URL? {
        let name = "txqr-\(Int(Date().timeIntervalSince1970)).txt"
        let url = FileManager.default.temporaryDirectory.appendingPathComponent(name)
        do {
            try recoveredText.write(to: url, atomically: true, encoding: .utf8)
            return url
        } catch {
            statusText = "Couldn't write file: \(error.localizedDescription)"
            return nil
        }
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
        leftActive = false
        rightActive = false
        integrityRetry = false
        lastProgressHaptic = 0
        statusText = "Scanning — LEFT/RIGHT dual or single auto"
        TransferLiveActivityStub.start(status: statusText)
    }
}
