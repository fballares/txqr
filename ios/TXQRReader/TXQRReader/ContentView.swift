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
                sideBadges
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
                    HStack {
                        Text("Transfer complete")
                            .font(.headline)
                            .foregroundStyle(.green)
                        Spacer()
                        Label("CRC verified", systemImage: "checkmark.seal.fill")
                            .font(.caption.weight(.semibold))
                            .foregroundStyle(.green)
                    }
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
                Text("Windows places dual codes LEFT and RIGHT. This app sorts detections the same way and merges both into one transfer.")
                    .font(.footnote)
                    .foregroundStyle(.secondary)
            }
        }
        .padding(16)
        .background(.ultraThinMaterial)
    }

    private var sideBadges: some View {
        HStack(spacing: 6) {
            sideChip("L", active: session.leftActive)
            sideChip("R", active: session.rightActive)
        }
    }

    private func sideChip(_ title: String, active: Bool) -> some View {
        Text(title)
            .font(.caption.weight(.bold).monospaced())
            .frame(width: 28, height: 28)
            .background(active ? Color.green.opacity(0.35) : Color.white.opacity(0.08), in: Circle())
            .foregroundStyle(active ? Color.green : Color.secondary)
    }
}
