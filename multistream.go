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
// Because each frame is independently useful to the LT decoder, reading two
// QRs in one camera exposure simply delivers two blocks per tick.

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
	// Spread streams evenly around the looping sequence so concurrent
	// cameras see distinct blocks as often as possible.
	offset := (frameCount * stream) / streams
	return (tick + offset) % frameCount
}

// SuggestedStreams returns how many concurrent QR slots to show for a
// payload size. Two is the practical phone-camera sweet spot.
func SuggestedStreams(payloadBytes int) int {
	if payloadBytes <= 0 {
		return 1
	}
	// Single QR is enough for short clipboard pastes.
	if payloadBytes <= DefaultClipboardProfile().ChunkLen {
		return 1
	}
	return 2
}
