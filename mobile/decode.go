package txqr // package should have this name to work properly with gomobile

import (
	"fmt"
	"time"

	"github.com/divan/txqr"
	"github.com/pyk/byten"
)

// Decoder implements txqr wrapper around protocol decoder.
type Decoder struct {
	*txqr.Decoder

	progress int
	speed    int // avg reading speed
	start    time.Time

	lastChunk    time.Time // last chunk decode request
	readInterval time.Duration
}

// NewDecoder creats new txqr decoder.
func NewDecoder() *Decoder {
	return &Decoder{
		Decoder: txqr.NewDecoder(),
	}
}

// Decode takes a single chunk of data and decodes it.
func (d *Decoder) Decode(data string) error {
	// mobile app can still try to decode any detected QR codes,
	// so we're ignoring them here
	if d.IsCompleted() {
		return nil
	}

	if err := d.Validate(data); err != nil {
		return err
	}

	// mark start of first txqr frame
	if d.start.IsZero() {
		d.start = time.Now()
	}

	// mark last chunk timestamp
	if !d.lastChunk.IsZero() {
		d.readInterval = time.Now().Sub(d.lastChunk)
	}
	d.lastChunk = time.Now()

	// decode
	err := d.Decoder.Decode(data)
	if err != nil {
		return err
	}

	d.updateProgress()

	return nil
}

// DecodeBatch ingests multiple QR payloads from one camera frame.
// Pass payloads joined with newlines (gomobile-friendly). Invalid or
// non-TXQR values are skipped so Vision multi-detect noise is safe.
// Returns how many payloads were accepted into the fountain decoder.
func (d *Decoder) DecodeBatch(joined string) int {
	if d.IsCompleted() || joined == "" {
		return 0
	}
	accepted := 0
	start := 0
	for i := 0; i <= len(joined); i++ {
		if i < len(joined) && joined[i] != '\n' {
			continue
		}
		part := joined[start:i]
		start = i + 1
		if part == "" {
			continue
		}
		if err := d.Decode(part); err != nil {
			// Non-TXQR codes in the scene are ignored.
			continue
		}
		accepted++
		if d.IsCompleted() {
			break
		}
	}
	return accepted
}

// UniqueFrames returns how many distinct TXQR frame headers have been seen.
func (d *Decoder) UniqueFrames() int {
	return d.UniqueCount()
}

// Speed returns avg reading speed.
func (d *Decoder) Speed() string {
	return fmt.Sprintf("%s/s", byten.Size(int64(d.speed)))
}

// ReadInterval returns the latest read interval in ms.
func (d *Decoder) ReadInterval() int64 {
	return int64(d.readInterval / time.Millisecond)
}

// TotalTime returns the total scan duration in human readable form - from first to last read chunk.
func (d *Decoder) TotalTime() string {
	dur := time.Since(d.start)
	return formatDuration(dur)
}

// formatDuration converts "12.232312313s" to "12.2s"
func formatDuration(d time.Duration) string {
	if d > time.Second {
		d = d - d%(100*time.Millisecond)
	}
	return d.String()
}

// TotalSize returns the data size in human readable form.
func (d *Decoder) TotalSize() string {
	return byten.Size(int64(d.Length()))
}

// Progress returns reading progress in percentage.
func (d *Decoder) Progress() int {
	return d.progress
}

// updateProgress updates progress and complete state of reading.
func (d *Decoder) updateProgress() {
	total := d.Total()
	if total <= 0 {
		d.progress = 0
		d.speed = 0
		return
	}
	read := d.Read()
	elapsed := time.Since(d.start)
	if elapsed > 0 {
		d.speed = read * int(time.Second) / int(elapsed)
	}
	d.progress = 100 * read / total
	if d.IsCompleted() {
		d.progress = 100
	}
}

// TotalTimeMs returns the total scan duration in milliseconds.
func (d *Decoder) TotalTimeMs() int64 {
	if d.start.IsZero() {
		return 0
	}
	dur := time.Since(d.start)
	return int64(dur / time.Millisecond)
}

// Reset resets decoder, preparing it for the next run.
func (d *Decoder) Reset() {
	d.Decoder.Reset()

	d.progress = 0
	d.speed = 0
	d.start = time.Time{}
	d.lastChunk = time.Time{}
	d.readInterval = 0
}
