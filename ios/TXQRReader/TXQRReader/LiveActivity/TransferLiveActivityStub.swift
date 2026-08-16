import Foundation

/// Placeholder for a future ActivityKit Live Activity / Dynamic Island update.
/// Ships as a no-op stub so the app builds without an ActivityKit entitlement
/// while keeping the call sites ready for a real widget extension later.
enum TransferLiveActivityStub {
    private static var lastProgress: Int = -1
    private static var active = false

    static func start(status: String) {
        active = true
        lastProgress = 0
        #if DEBUG
        print("[TXQR LiveActivity stub] start: \(status)")
        #endif
    }

    static func update(progress: Int, status: String) {
        guard active else { return }
        if progress == lastProgress { return }
        lastProgress = progress
        #if DEBUG
        print("[TXQR LiveActivity stub] \(progress)% — \(status)")
        #endif
    }

    static func end() {
        active = false
        lastProgress = -1
        #if DEBUG
        print("[TXQR LiveActivity stub] end")
        #endif
    }
}
