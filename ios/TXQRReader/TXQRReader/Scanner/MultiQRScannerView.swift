import SwiftUI
import AVFoundation
import Vision

/// Live camera preview that runs Vision QR detection and forwards *all*
/// QR payloads in each frame to `TransferSession` for concurrent multi-QR ingest.
struct MultiQRScannerView: UIViewControllerRepresentable {
    @ObservedObject var session: TransferSession

    func makeUIViewController(context: Context) -> MultiQRScannerController {
        let controller = MultiQRScannerController()
        controller.onCodes = { codes in
            Task { @MainActor in
                session.ingest(codes: codes)
            }
        }
        return controller
    }

    func updateUIViewController(_ uiViewController: MultiQRScannerController, context: Context) {}
}

final class MultiQRScannerController: UIViewController, AVCaptureVideoDataOutputSampleBufferDelegate {
    var onCodes: (([String]) -> Void)?

    private let session = AVCaptureSession()
    private let videoOutput = AVCaptureVideoDataOutput()
    private let visionQueue = DispatchQueue(label: "txqr.vision", qos: .userInitiated)
    private var previewLayer: AVCaptureVideoPreviewLayer?
    private var lastEmit = Date.distantPast
    private let minInterval: TimeInterval = 1.0 / 12.0 // cap Vision work ~12 Hz

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

        // Prefer continuous autofocus for screen-to-camera QR.
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
            let payloads = results.compactMap { obs -> String? in
                guard obs.symbology == .qr, let payload = obs.payloadStringValue, !payload.isEmpty else {
                    return nil
                }
                return payload
            }
            guard !payloads.isEmpty else { return }
            // Dedupe identical payloads within the same frame.
            let unique = Array(Set(payloads))
            DispatchQueue.main.async {
                self?.onCodes?(unique)
            }
        }
        request.symbologies = [.qr]

        let handler = VNImageRequestHandler(cvPixelBuffer: pixelBuffer, orientation: .right, options: [:])
        try? handler.perform([request])
    }
}
