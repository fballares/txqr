import SwiftUI
import AVFoundation
import Vision

/// One QR payload with its horizontal position in the upright image (0…1).
struct PositionedQR: Equatable {
    let payload: String
    /// Normalized mid-X after Vision orientation correction (0 = left, 1 = right).
    let midX: CGFloat
}

/// Live camera preview: Vision multi-QR detect, sorted LEFT → RIGHT for dual slots.
struct MultiQRScannerView: UIViewControllerRepresentable {
    @ObservedObject var session: TransferSession

    func makeUIViewController(context: Context) -> MultiQRScannerController {
        let controller = MultiQRScannerController()
        controller.onCodes = { positioned in
            Task { @MainActor in
                session.ingest(positioned: positioned)
            }
        }
        return controller
    }

    func updateUIViewController(_ uiViewController: MultiQRScannerController, context: Context) {}
}

final class MultiQRScannerController: UIViewController, AVCaptureVideoDataOutputSampleBufferDelegate {
    var onCodes: (([PositionedQR]) -> Void)?

    private let session = AVCaptureSession()
    private let videoOutput = AVCaptureVideoDataOutput()
    private let visionQueue = DispatchQueue(label: "txqr.vision", qos: .userInitiated)
    private var previewLayer: AVCaptureVideoPreviewLayer?
    private var lastEmit = Date.distantPast
    private let minInterval: TimeInterval = 1.0 / 12.0

    override func viewDidLoad() {
        super.viewDidLoad()
        view.backgroundColor = .black
        configureSession()
    }

    override func viewDidLayoutSubviews() {
        super.viewDidLayoutSubviews()
        previewLayer?.frame = view.bounds
    }

    override func viewWillAppear(_ animated: Bool) {
        super.viewWillAppear(animated)
        visionQueue.async { [weak self] in
            guard let self, !self.session.isRunning else { return }
            self.session.startRunning()
        }
    }

    override func viewWillDisappear(_ animated: Bool) {
        super.viewWillDisappear(animated)
        visionQueue.async { [weak self] in
            self?.session.stopRunning()
        }
    }

    private func configureSession() {
        session.beginConfiguration()
        session.sessionPreset = .high

        guard
            let device = AVCaptureDevice.default(.builtInWideAngleCamera, for: .video, position: .back),
            let input = try? AVCaptureDeviceInput(device: device),
            session.canAddInput(input)
        else {
            return
        }
        session.addInput(input)

        try? device.lockForConfiguration()
        if device.isFocusModeSupported(.continuousAutoFocus) {
            device.focusMode = .continuousAutoFocus
        }
        if device.isExposureModeSupported(.continuousAutoExposure) {
            device.exposureMode = .continuousAutoExposure
        }
        device.unlockForConfiguration()

        videoOutput.alwaysDiscardsLateVideoFrames = true
        videoOutput.setSampleBufferDelegate(self, queue: visionQueue)
        guard session.canAddOutput(videoOutput) else { return }
        session.addOutput(videoOutput)
        if let connection = videoOutput.connection(with: .video), connection.isVideoStabilizationSupported {
            connection.preferredVideoStabilizationMode = .off
        }

        session.commitConfiguration()

        let preview = AVCaptureVideoPreviewLayer(session: session)
        preview.videoGravity = .resizeAspectFill
        preview.frame = view.bounds
        view.layer.addSublayer(preview)
        previewLayer = preview
    }

    func captureOutput(
        _ output: AVCaptureOutput,
        didOutput sampleBuffer: CMSampleBuffer,
        from connection: AVCaptureConnection
    ) {
        let now = Date()
        guard now.timeIntervalSince(lastEmit) >= minInterval else { return }
        lastEmit = now

        guard let pixelBuffer = CMSampleBufferGetImageBuffer(sampleBuffer) else { return }

        let request = VNDetectBarcodesRequest { [weak self] req, error in
            guard error == nil, let results = req.results as? [VNBarcodeObservation] else { return }

            // Keep position so we can map to Windows LEFT / RIGHT slots.
            var positioned: [PositionedQR] = []
            positioned.reserveCapacity(results.count)
            var seenPayload = Set<String>()
            for obs in results where obs.symbology == .qr {
                guard let payload = obs.payloadStringValue, !payload.isEmpty else { continue }
                if seenPayload.contains(payload) { continue }
                seenPayload.insert(payload)
                let midX = obs.boundingBox.midX
                positioned.append(PositionedQR(payload: payload, midX: midX))
            }
            guard !positioned.isEmpty else { return }

            // LEFT → RIGHT (stream 0, stream 1 on the Windows overlay).
            positioned.sort { $0.midX < $1.midX }

            DispatchQueue.main.async {
                self?.onCodes?(positioned)
            }
        }
        request.symbologies = [.qr]

        // .right matches typical portrait back-camera buffers so midX is left→right
        // in the upright framed image the user sees.
        let handler = VNImageRequestHandler(cvPixelBuffer: pixelBuffer, orientation: .right, options: [:])
        try? handler.perform([request])
    }
}
