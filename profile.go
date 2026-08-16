package txqr

// ClipboardProfile holds encoder/animation defaults tuned for
// screen-to-phone clipboard text transfer on Windows.
//
// Rationale (phone camera aimed at a bright monitor):
//   - Medium ECC: enough resilience to focus/angle noise without
//     bloating each frame the way High/Highest do.
//   - Chunk ~100-160 bytes: stays in mid QR versions that decode
//     reliably at ~0.5m viewing distance.
//   - 5-6 FPS: slow enough for typical mobile decode pipelines;
//     faster rates drop frames more often than they help.
//   - Redundancy 2.0-2.5: fountain codes recover from missed frames
//     without making the loop excessively long for clipboard-sized text.
type ClipboardProfile struct {
	ChunkLen   int
	FPS        int
	QRSize     int
	Redundancy float64
}

// DefaultClipboardProfile returns baseline settings for typical
// copied text (notes, URLs, email drafts, code snippets).
func DefaultClipboardProfile() ClipboardProfile {
	return ClipboardProfile{
		ChunkLen:   140,
		FPS:        6,
		QRSize:     560,
		Redundancy: 2.0,
	}
}

// ProfileForPayload chooses settings based on payload size.
// Larger pastes use slightly smaller chunks, lower FPS, and more
// redundancy so a handheld phone can still finish a loop.
func ProfileForPayload(n int) ClipboardProfile {
	p := DefaultClipboardProfile()
	switch {
	case n <= 0:
		return p
	case n <= p.ChunkLen:
		// Fits in one frame — animation settings unused.
		return p
	case n <= 2*1024:
		p.ChunkLen = 150
		p.FPS = 6
		p.Redundancy = 2.0
	case n <= 8*1024:
		p.ChunkLen = 130
		p.FPS = 5
		p.Redundancy = 2.25
	default:
		// Long pastes: prioritize reliable decode over speed.
		p.ChunkLen = 110
		p.FPS = 5
		p.Redundancy = 2.5
		p.QRSize = 640
	}
	return p
}

// SuggestChunkLen returns the chunk length ProfileForPayload would use.
func SuggestChunkLen(n int) int {
	return ProfileForPayload(n).ChunkLen
}
