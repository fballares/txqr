package txqr

import "testing"

func TestProfileForPayload(t *testing.T) {
	small := ProfileForPayload(40)
	if small.ChunkLen < 40 {
		t.Fatalf("small payload chunk %d should cover the text", small.ChunkLen)
	}

	medium := ProfileForPayload(3000)
	if medium.ChunkLen >= small.ChunkLen && medium.FPS > small.FPS {
		t.Fatalf("medium profile should not be more aggressive than small: %+v vs %+v", medium, small)
	}

	large := ProfileForPayload(20 * 1024)
	if large.Redundancy < medium.Redundancy {
		t.Fatalf("large profile should increase redundancy: %+v", large)
	}
	if large.FPS > 5 {
		t.Fatalf("large profile FPS too high for reliable phone decode: %d", large.FPS)
	}
}

func TestSuggestChunkLen(t *testing.T) {
	if got := SuggestChunkLen(100); got != ProfileForPayload(100).ChunkLen {
		t.Fatalf("SuggestChunkLen mismatch: %d", got)
	}
}
