import SwiftUI

struct ContentView: View {
    @StateObject private var session = TransferSession()

    var body: some View {
        ZStack {
            Color.black.ignoresSafeArea()

            VStack(spacing: 0) {
                MultiQRScannerView(session: session)
                    .ignoresSafeArea(edges: .top)

                statusPanel
            }
        }
        .onAppear { session.start() }
        .onDisappear { session.stop() }
    }

    private var statusPanel: some View {
        VStack(alignment: .leading, spacing: 12) {
            HStack {
                Text("TXQR Reader")
                    .font(.title2.weight(.bold))
                Spacer()
                layoutBadge
            }

            ProgressView(value: Double(session.progress), total: 100)
                .tint(.green)

            HStack {
                Text(session.statusText)
                    .font(.footnote)
                    .foregroundStyle(.secondary)
                Spacer()
                Text("\(session.progress)%")
                    .font(.footnote.monospacedDigit())
            }

            if session.isComplete {
                VStack(alignment: .leading, spacing: 8) {
                    Text("Transfer complete")
                        .font(.headline)
                        .foregroundStyle(.green)
                    Text(session.preview)
                        .font(.system(.footnote, design: .monospaced))
                        .lineLimit(6)
                        .frame(maxWidth: .infinity, alignment: .leading)
                        .padding(10)
                        .background(Color.white.opacity(0.06), in: RoundedRectangle(cornerRadius: 8))

                    HStack {
                        Button("Copy") { session.copyToClipboard() }
                            .buttonStyle(.borderedProminent)
                        ShareLink(item: session.recoveredText) {
                            Label("Share", systemImage: "square.and.arrow.up")
                        }
                        .buttonStyle(.bordered)
                        Button("Scan again") { session.reset() }
                            .buttonStyle(.bordered)
                    }
                }
            } else {
                Text("Windows auto-picks single or dual QR for speed. This app reads whatever is on screen — 1 or 2 — in the same scan loop.")
                    .font(.footnote)
                    .foregroundStyle(.secondary)
            }
        }
        .padding(16)
        .background(.ultraThinMaterial)
    }

    private var layoutBadge: some View {
        let dual = session.detectedLayout == .dual || session.detectedLayout == .multi
        return Label(
            "\(session.lastCodesInFrame) now · peak \(session.peakConcurrent)",
            systemImage: dual ? "square.grid.2x2" : "qrcode.viewfinder"
        )
        .font(.caption.monospacedDigit())
        .padding(.horizontal, 8)
        .padding(.vertical, 4)
        .background(dual ? Color.green.opacity(0.25) : Color.white.opacity(0.08), in: Capsule())
    }
}
