import SwiftUI
import AVFoundation

struct ContentView: View {
    @StateObject private var session = TransferSession()
    @StateObject private var camera = CameraControls()
    @State private var showCoach = !UserDefaults.standard.bool(forKey: "txqr_coach_done")
    @State private var showSuccess = false
    @State private var saveURL: URL?

    var body: some View {
        ZStack {
            Color(red: 0.03, green: 0.04, blue: 0.05).ignoresSafeArea()

            VStack(spacing: 0) {
                ZStack {
                    MultiQRScannerView(session: session, camera: camera)
                        .ignoresSafeArea(edges: .top)

                    AimBracketsOverlay(
                        dualHint: session.detectedLayout == .dual || session.peakConcurrent >= 2,
                        leftActive: session.leftActive,
                        rightActive: session.rightActive
                    )
                    .allowsHitTesting(false)

                    VStack {
                        topBar
                        Spacer()
                        progressRing
                            .padding(.bottom, 18)
                    }
                    .padding(16)
                }
                .frame(maxHeight: .infinity)

                statusPanel
            }

            if showCoach {
                CoachOverlay {
                    UserDefaults.standard.set(true, forKey: "txqr_coach_done")
                    showCoach = false
                }
            }
        }
        .preferredColorScheme(.dark)
        .onAppear {
            session.start()
            UIApplication.shared.isIdleTimerDisabled = true
        }
        .onDisappear {
            session.stop()
            UIApplication.shared.isIdleTimerDisabled = false
        }
        .onChange(of: session.isComplete) { done in
            if done {
                showSuccess = true
                TransferLiveActivityStub.update(progress: 100, status: "Complete")
            } else {
                TransferLiveActivityStub.update(progress: session.progress, status: session.statusText)
            }
        }
        .onChange(of: session.progress) { value in
            TransferLiveActivityStub.update(progress: value, status: session.statusText)
        }
        .sheet(isPresented: $showSuccess) {
            SuccessSheet(
                text: session.recoveredText,
                preview: session.preview,
                status: session.statusText,
                onCopy: {
                    session.copyToClipboard()
                    session.playCopyHaptic()
                },
                onSave: {
                    saveURL = session.writeTempFile()
                },
                onScanAgain: {
                    showSuccess = false
                    session.reset()
                }
            )
            .presentationDetents([.medium, .large])
        }
        .sheet(item: Binding(
            get: { saveURL.map(IdentifiableURL.init) },
            set: { saveURL = $0?.url }
        )) { item in
            ShareSheet(items: [item.url])
        }
    }

    private var topBar: some View {
        HStack(spacing: 10) {
            Text("txqr")
                .font(.system(.title3, design: .monospaced).weight(.bold))
                .foregroundStyle(Color(red: 0.95, green: 0.94, blue: 0.92))

            Spacer()

            sideChip("L", active: session.leftActive)
            sideChip("R", active: session.rightActive)

            Button {
                camera.toggleTorch()
            } label: {
                Image(systemName: camera.torchOn ? "flashlight.on.fill" : "flashlight.off.fill")
                    .font(.body.weight(.semibold))
                    .foregroundStyle(camera.torchOn ? Color(red: 0.88, green: 0.48, blue: 0.24) : .white)
                    .frame(width: 36, height: 36)
                    .background(.black.opacity(0.35), in: Circle())
            }
            .accessibilityLabel("Toggle torch")
        }
    }

    private var progressRing: some View {
        ZStack {
            Circle()
                .stroke(Color.white.opacity(0.12), lineWidth: 10)
            Circle()
                .trim(from: 0, to: CGFloat(session.progress) / 100)
                .stroke(
                    Color(red: 0.88, green: 0.48, blue: 0.24),
                    style: StrokeStyle(lineWidth: 10, lineCap: .round)
                )
                .rotationEffect(.degrees(-90))
                .animation(.easeInOut(duration: 0.2), value: session.progress)
            VStack(spacing: 2) {
                Text("\(session.progress)%")
                    .font(.system(.title2, design: .monospaced).weight(.bold))
                    .foregroundStyle(.white)
                Text(session.detectedLayout.rawValue)
                    .font(.caption2.weight(.semibold))
                    .foregroundStyle(.white.opacity(0.7))
            }
        }
        .frame(width: 118, height: 118)
        .padding(10)
        .background(.black.opacity(0.28), in: Circle())
    }

    private var statusPanel: some View {
        VStack(alignment: .leading, spacing: 10) {
            Text(session.statusText)
                .font(.footnote)
                .foregroundStyle(Color(red: 0.55, green: 0.58, blue: 0.62))
                .frame(maxWidth: .infinity, alignment: .leading)

            if !session.isComplete {
                Text("Windows LEFT → phone L · Windows RIGHT → phone R. Keep both codes inside the brackets.")
                    .font(.caption)
                    .foregroundStyle(Color(red: 0.55, green: 0.58, blue: 0.62))
            } else {
                HStack {
                    Button("Copy") {
                        session.copyToClipboard()
                        session.playCopyHaptic()
                    }
                    .buttonStyle(.borderedProminent)
                    .tint(Color(red: 0.88, green: 0.48, blue: 0.24))

                    Button("Save file") {
                        saveURL = session.writeTempFile()
                    }
                    .buttonStyle(.bordered)

                    Button("Scan again") {
                        session.reset()
                        showSuccess = false
                    }
                    .buttonStyle(.bordered)
                }
            }
        }
        .padding(16)
        .background(Color(red: 0.07, green: 0.09, blue: 0.11))
    }

    private func sideChip(_ title: String, active: Bool) -> some View {
        Text(title)
            .font(.caption.weight(.bold).monospaced())
            .frame(width: 28, height: 28)
            .background(
                active
                    ? Color(red: 0.88, green: 0.48, blue: 0.24).opacity(0.45)
                    : Color.white.opacity(0.08),
                in: Circle()
            )
            .foregroundStyle(active ? Color(red: 0.98, green: 0.9, blue: 0.8) : .secondary)
    }
}

private struct IdentifiableURL: Identifiable {
    let id = UUID()
    let url: URL
}

struct AimBracketsOverlay: View {
    var dualHint: Bool
    var leftActive: Bool
    var rightActive: Bool

    var body: some View {
        GeometryReader { geo in
            let padX = geo.size.width * 0.06
            let padY = geo.size.height * 0.12
            let gap = dualHint ? geo.size.width * 0.04 : 0
            let boxW = dualHint
                ? (geo.size.width - padX * 2 - gap) / 2
                : geo.size.width - padX * 2
            let boxH = geo.size.height - padY * 2

            HStack(spacing: gap) {
                BracketFrame(active: leftActive || !dualHint, label: dualHint ? "LEFT" : nil)
                    .frame(width: boxW, height: boxH)
                if dualHint {
                    BracketFrame(active: rightActive, label: "RIGHT")
                        .frame(width: boxW, height: boxH)
                }
            }
            .padding(.horizontal, padX)
            .padding(.vertical, padY)
            .frame(maxWidth: .infinity, maxHeight: .infinity)
        }
    }
}

struct BracketFrame: View {
    var active: Bool
    var label: String?

    var body: some View {
        ZStack(alignment: .top) {
            RoundedRectangle(cornerRadius: 8)
                .stroke(
                    active ? Color(red: 0.88, green: 0.48, blue: 0.24) : Color.white.opacity(0.35),
                    style: StrokeStyle(lineWidth: active ? 3 : 2, dash: active ? [] : [6, 5])
                )
            if let label {
                Text(label)
                    .font(.caption2.weight(.bold).monospaced())
                    .tracking(1.5)
                    .foregroundStyle(Color(red: 0.88, green: 0.48, blue: 0.24))
                    .padding(.horizontal, 8)
                    .padding(.vertical, 4)
                    .background(.black.opacity(0.45), in: Capsule())
                    .offset(y: -14)
            }
        }
    }
}

struct CoachOverlay: View {
    var onDismiss: () -> Void

    var body: some View {
        ZStack {
            Color.black.opacity(0.72).ignoresSafeArea()
            VStack(alignment: .leading, spacing: 16) {
                Text("txqr")
                    .font(.system(.largeTitle, design: .monospaced).weight(.bold))
                Text("Aim at the Windows overlay. Dual mode shows LEFT and RIGHT codes — keep both in frame. Torch helps in bright rooms.")
                    .foregroundStyle(.secondary)
                VStack(alignment: .leading, spacing: 8) {
                    Label("Copy text on Windows, press the hotkey", systemImage: "1.circle.fill")
                    Label("Point the camera at the floating QR(s)", systemImage: "2.circle.fill")
                    Label("When the ring hits 100%, Copy or Save", systemImage: "3.circle.fill")
                }
                .font(.subheadline)
                Button("Got it") { onDismiss() }
                    .buttonStyle(.borderedProminent)
                    .tint(Color(red: 0.88, green: 0.48, blue: 0.24))
                    .padding(.top, 8)
            }
            .padding(24)
            .frame(maxWidth: 360)
            .background(Color(red: 0.1, green: 0.12, blue: 0.14), in: RoundedRectangle(cornerRadius: 16))
        }
    }
}

struct SuccessSheet: View {
    let text: String
    let preview: String
    let status: String
    var onCopy: () -> Void
    var onSave: () -> Void
    var onScanAgain: () -> Void

    var body: some View {
        NavigationStack {
            VStack(alignment: .leading, spacing: 14) {
                Label("Transfer complete", systemImage: "checkmark.seal.fill")
                    .font(.title2.weight(.bold))
                    .foregroundStyle(Color(red: 0.24, green: 0.72, blue: 0.48))
                Text(status)
                    .font(.footnote)
                    .foregroundStyle(.secondary)
                Text(preview)
                    .font(.system(.footnote, design: .monospaced))
                    .padding(12)
                    .frame(maxWidth: .infinity, alignment: .leading)
                    .background(Color.secondary.opacity(0.12), in: RoundedRectangle(cornerRadius: 10))
                HStack {
                    Button("Copy", action: onCopy)
                        .buttonStyle(.borderedProminent)
                        .tint(Color(red: 0.88, green: 0.48, blue: 0.24))
                    Button("Save to Files", action: onSave)
                        .buttonStyle(.bordered)
                    Button("Scan again", action: onScanAgain)
                        .buttonStyle(.bordered)
                }
                Spacer()
            }
            .padding(20)
            .navigationTitle("txqr")
            .navigationBarTitleDisplayMode(.inline)
        }
    }
}

struct ShareSheet: UIViewControllerRepresentable {
    let items: [Any]
    func makeUIViewController(context: Context) -> UIActivityViewController {
        UIActivityViewController(activityItems: items, applicationActivities: nil)
    }
    func updateUIViewController(_ uiViewController: UIActivityViewController, context: Context) {}
}

/// Torch + exposure helpers shared with the scanner controller.
@MainActor
final class CameraControls: ObservableObject {
    @Published var torchOn = false
    weak var device: AVCaptureDevice?

    func toggleTorch() {
        guard let device, device.hasTorch else { return }
        do {
            try device.lockForConfiguration()
            let next = !torchOn
            if next {
                try device.setTorchModeOn(level: 0.75)
            } else {
                device.torchMode = .off
            }
            torchOn = next
            device.unlockForConfiguration()
        } catch {
            torchOn = false
        }
    }

    func bind(device: AVCaptureDevice?) {
        self.device = device
    }
}
