package txqr

// Multi-QR / concurrent streams
//
// TXQR fountain frames already interoperate with multiple on-screen QR codes
// without changing the wire format. The sender shows different frames from the
// same Encode() result in each display slot; the phone decodes every visible
// QR in a camera frame and feeds all payloads into one Decoder.
//
// Frame format (unchanged):
//
//	blockCode/chunkLen/total|payload
//
// Dual QR is only used when it is likely to finish faster. Tiny pastes stay on
// one larger code so the phone can lock focus more easily.

const (
	// minBytesForDual: below this, a single QR (or a short loop) is enough.
	minBytesForDual = 2 * 1024
	// minFramesForDual: need enough distinct frames for parallelism to matter.
	minFramesForDual = 12
	// minLoopSecondsForDual: if one full animation loop is shorter than this,
	// dual QR adds optical complexity without a meaningful speedup.
	minLoopSecondsForDual = 2.0
)

// FrameIndexForStream picks which encoded frame a display slot should show.
// tick advances once per animation step; stream is [0, streams).
func FrameIndexForStream(tick, stream, streams, frameCount int) int {
	if frameCount <= 0 {
		return 0
	}
	if streams < 1 {
		streams = 1
	}
	if stream < 0 {
		stream = 0
	}
	if stream >= streams {
		stream %= streams
	}
	offset := (frameCount * stream) / streams
	return (tick + offset) % frameCount
}

// SuggestedStreams estimates 1 vs 2 QR slots from payload size alone
// (before encoding). Prefer SuggestedStreamsForTransfer after Encode.
func SuggestedStreams(payloadBytes int) int {
	if payloadBytes <= 0 {
		return 1
	}
	p := ProfileForPayload(payloadBytes)
	if payloadBytes <= p.ChunkLen {
		return 1
	}
	nChunks := numberOfChunks(payloadBytes, p.ChunkLen)
	frames := int(float64(nChunks) * p.Redundancy)
	if frames < 1 {
		frames = 1
	}
	return SuggestedStreamsForTransfer(payloadBytes, frames, p.FPS)
}

// SuggestedStreamsForTransfer chooses concurrent QR count after encoding.
// Returns 2 only when dual display should reduce wall-clock transfer time
// enough to justify smaller on-screen codes.
func SuggestedStreamsForTransfer(payloadBytes, frameCount, fps int) int {
	if frameCount <= 1 {
		return 1
	}
	if payloadBytes < minBytesForDual {
		return 1
	}
	if frameCount < minFramesForDual {
		return 1
	}
	if fps < 1 {
		fps = 6
	}
	loopSec := float64(frameCount) / float64(fps)
	if loopSec < minLoopSecondsForDual {
		return 1
	}
	return 2
}

// OverlaySize suggests pixel size for the sender window given stream count.
func OverlaySize(streams int) (width, height int) {
	if streams >= 2 {
		return 720, 500
	}
	return 420, 540
}
