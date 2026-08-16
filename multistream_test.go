package txqr

import "testing"

func TestFrameIndexForStreamSpreads(t *testing.T) {
	n := 10
	seen := map[int]bool{}
	for s := 0; s < 2; s++ {
		idx := FrameIndexForStream(0, s, 2, n)
		seen[idx] = true
	}
	if len(seen) != 2 {
		t.Fatalf("expected two distinct frames at tick 0, got %v", seen)
	}
}

func TestFrameIndexWraps(t *testing.T) {
	if got := FrameIndexForStream(11, 0, 1, 10); got != 1 {
		t.Fatalf("got %d", got)
	}
}

func TestSuggestedStreams(t *testing.T) {
	if SuggestedStreams(20) != 1 {
		t.Fatal("short payload should be 1 stream")
	}
	if SuggestedStreams(5000) != 2 {
		t.Fatal("longer payload should suggest 2 streams")
	}
}
