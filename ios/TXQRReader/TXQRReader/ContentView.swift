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
                Label("\(session.lastCodesInFrame) QR", systemImage: "qrcode.viewfinder")
                    .font(.subheadline.monospacedDigit())
                    .foregroundStyle(.secondary)
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
                Text("Point at the Windows overlay. Multiple QR codes in one view are decoded together. You can start mid-loop.")
                    .font(.footnote)
                    .foregroundStyle(.secondary)
            }
        }
        .padding(16)
        .background(.ultraThinMaterial)
    }
}
