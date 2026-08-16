package txqr

import (
	"fmt"
	"hash/crc32"
	"math/rand"
	"strings"

	fountain "github.com/google/gofountain"
)

// Decoder represents protocol decode.
type Decoder struct {
	chunkLen  int
	codec     fountain.Codec
	fd        fountain.Decoder
	completed bool
	verified  bool
	total     int
	expectCRC uint32
	hasCRC    bool
	cache     map[string]struct{}
	crcErr    error
}

// NewDecoder creates and inits a new decoder.
func NewDecoder() *Decoder {
	return &Decoder{
		cache: make(map[string]struct{}),
	}
}

// NewDecoderSize creates and inits a new decoder for the known size.
func NewDecoderSize(size, chunkLen int) *Decoder {
	numChunks := numberOfChunks(size, chunkLen)
	codec := fountain.NewLubyCodec(numChunks, rand.New(fountain.NewMersenneTwister(200)), solitonDistribution(numChunks))
	return &Decoder{
		codec:    codec,
		fd:       codec.NewDecoder(size),
		total:    size,
		chunkLen: chunkLen,
		cache:    make(map[string]struct{}),
	}
}

// Decode takes a single chunk of data and decodes it.
// Chunk expected to be validated (see Validate) before.
func (d *Decoder) Decode(chunk string) error {
	idx := strings.IndexByte(chunk, '|') // expected to be validated before
	if idx == -1 {
		return fmt.Errorf("invalid frame: %q", chunk)
	}

	header := chunk[:idx]
	// continuous QR reading often sends the same chunk in a row, skip it
	if d.isCached(header) {
		return nil
	}

	blockCode, chunkLen, total, crc, hasCRC, err := parseHeader(header)
	if err != nil {
		return fmt.Errorf("invalid header: %v (%s)", err, header)
	}

	payload := chunk[idx+1:]
	lubyBlock := fountain.LTBlock{
		BlockCode: blockCode,
		Data:      []byte(payload),
	}

	if d.fd == nil {
		d.total = total
		d.chunkLen = chunkLen
		d.expectCRC = crc
		d.hasCRC = hasCRC
		numChunks := numberOfChunks(d.total, d.chunkLen)
		d.codec = fountain.NewLubyCodec(numChunks, rand.New(fountain.NewMersenneTwister(200)), solitonDistribution(numChunks))
		d.fd = d.codec.NewDecoder(total)
	} else if hasCRC {
		// All frames of a transfer must advertise the same payload CRC.
		if !d.hasCRC {
			d.hasCRC = true
			d.expectCRC = crc
		} else if crc != d.expectCRC {
			return fmt.Errorf("CRC mismatch across frames: got %08x want %08x", crc, d.expectCRC)
		}
	}

	done := d.fd.AddBlocks([]fountain.LTBlock{lubyBlock})
	if done {
		if err := d.finalize(); err != nil {
			// Allow the next animation loop to re-feed frames after a bad reconstruct.
			d.cache = map[string]struct{}{}
			return err
		}
	}
	return nil
}

func (d *Decoder) finalize() error {
	data := d.fd.Decode()
	if d.hasCRC {
		got := crc32.ChecksumIEEE(data)
		if got != d.expectCRC {
			d.crcErr = fmt.Errorf("payload CRC mismatch: got %08x want %08x", got, d.expectCRC)
			d.completed = false
			d.verified = false
			// Reset fountain state so the phone can keep scanning a fresh loop.
			d.fd = d.codec.NewDecoder(d.total)
			return d.crcErr
		}
	}
	d.completed = true
	d.verified = true
	d.crcErr = nil
	return nil
}

func parseHeader(header string) (blockCode int64, chunkLen, total int, crc uint32, hasCRC bool, err error) {
	// New format: blockCode/chunkLen/total/crc32
	var crcHex string
	n, scanErr := fmt.Sscanf(header, "%d/%d/%d/%s", &blockCode, &chunkLen, &total, &crcHex)
	if scanErr == nil && n == 4 {
		var parsed uint32
		_, err = fmt.Sscanf(crcHex, "%x", &parsed)
		if err != nil {
			return 0, 0, 0, 0, false, fmt.Errorf("bad CRC field %q", crcHex)
		}
		return blockCode, chunkLen, total, parsed, true, nil
	}
	// Legacy format without CRC (still accepted; verification skipped).
	n, scanErr = fmt.Sscanf(header, "%d/%d/%d", &blockCode, &chunkLen, &total)
	if scanErr != nil || n != 3 {
		return 0, 0, 0, 0, false, fmt.Errorf("parse failed")
	}
	return blockCode, chunkLen, total, 0, false, nil
}

// Validate checks if a given chunk of data is a valid txqr protocol packet.
func (d *Decoder) Validate(chunk string) error {
	if chunk == "" || len(chunk) < 4 {
		return fmt.Errorf("invalid frame: %q", chunk)
	}
	idx := strings.IndexByte(chunk, '|')
	if idx == -1 {
		return fmt.Errorf("invalid frame: %q", chunk)
	}
	return nil
}

// Data returns decoded data (only after a successful integrity check when CRC present).
func (d *Decoder) Data() string {
	return string(d.DataBytes())
}

// DataBytes returns decoded data as a byte slice.
func (d *Decoder) DataBytes() []byte {
	if d.fd == nil || !d.completed {
		return []byte{}
	}
	return d.fd.Decode()
}

// Length returns the expected total size of the decoded data in bytes.
func (d *Decoder) Length() int {
	return d.total
}

// Read returns an estimate of how many bytes have been received so far.
func (d *Decoder) Read() int {
	if d.completed {
		return d.total
	}
	if d.chunkLen <= 0 || d.total <= 0 {
		return 0
	}
	n := len(d.cache) * d.chunkLen
	if n > d.total {
		return d.total
	}
	return n
}

// Total returns total amount of data.
func (d *Decoder) Total() int {
	return d.total
}

// UniqueCount returns how many distinct frame headers have been observed.
func (d *Decoder) UniqueCount() int {
	return len(d.cache)
}

// ExpectedCRC returns the CRC-32 advertised by the sender (0 if unknown).
func (d *Decoder) ExpectedCRC() uint32 {
	return d.expectCRC
}

// ExpectedCRCHex returns the expected CRC as 8 lowercase hex digits, or "".
func (d *Decoder) ExpectedCRCHex() string {
	if !d.hasCRC {
		return ""
	}
	return fmt.Sprintf("%08x", d.expectCRC)
}

// HasCRC reports whether the transfer advertises a payload CRC.
func (d *Decoder) HasCRC() bool {
	return d.hasCRC
}

// IsVerified reports whether reconstructed data passed the CRC check.
func (d *Decoder) IsVerified() bool {
	return d.verified
}

// IntegrityError returns the last CRC failure, if any.
func (d *Decoder) IntegrityError() string {
	if d.crcErr == nil {
		return ""
	}
	return d.crcErr.Error()
}

// IsCompleted reports whether decoding finished and integrity checks passed.
func (d *Decoder) IsCompleted() bool {
	return d.completed && d.verified
}

// Reset resets decoder, preparing it for the next run.
func (d *Decoder) Reset() {
	d.fd = nil
	d.completed = false
	d.verified = false
	d.chunkLen = 0
	d.total = 0
	d.expectCRC = 0
	d.hasCRC = false
	d.crcErr = nil
	d.cache = map[string]struct{}{}
	d.codec = nil
}

func (d *Decoder) isCached(header string) bool {
	if _, ok := d.cache[header]; ok {
		return true
	}
	d.cache[header] = struct{}{}
	return false
}
