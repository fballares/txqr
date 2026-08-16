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

func TestSuggestedStreamsAutoBenefit(t *testing.T) {
	// Short clipboard → single
	if SuggestedStreams(100) != 1 {
		t.Fatal("tiny payload should be 1 stream")
	}
	if SuggestedStreamsForTransfer(500, 4, 6) != 1 {
		t.Fatal("few frames should stay single")
	}
	if SuggestedStreamsForTransfer(1500, 20, 6) != 1 {
		t.Fatal("under 2KB should stay single even with many frames")
	}
	// Larger body with a long enough loop → dual
	if SuggestedStreamsForTransfer(8*1024, 30, 6) != 2 {
		t.Fatal("8KB / 30 frames @6fps should use dual")
	}
	if SuggestedStreamsForTransfer(8*1024, 30, 20) != 1 {
		t.Fatal("fast short loop should stay single")
	}
}

func TestOverlaySize(t *testing.T) {
	w1, _ := OverlaySize(1)
	w2, _ := OverlaySize(2)
	if w2 <= w1 {
		t.Fatalf("dual overlay should be wider: %d vs %d", w2, w1)
	}
}
