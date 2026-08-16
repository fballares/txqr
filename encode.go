package txqr

import (
	"fmt"
	"hash/crc32"
	"math/rand"

	fountain "github.com/google/gofountain"
)

// Encoder represents protocol encoder.
type Encoder struct {
	chunkLen         int
	redundancyFactor float64
}

// NewEncoder creates and inits a new encoder for the given chunk length.
func NewEncoder(n int) *Encoder {
	return &Encoder{
		chunkLen:         n,
		redundancyFactor: 2.0,
	}
}

// Encode encodes data and splits it into QR frames.
// Each frame header carries a CRC-32 of the full payload so the receiver
// can verify end-to-end integrity after fountain reconstruction.
func (e *Encoder) Encode(str string) ([]string, error) {
	crc := crc32.ChecksumIEEE([]byte(str))

	// A single QR frame is optimal for short clipboard pastes.
	if len(str) <= e.chunkLen {
		return []string{e.frame(0, len(str), crc, []byte(str))}, nil
	}

	numChunks := numberOfChunks(len(str), e.chunkLen)
	codec := fountain.NewLubyCodec(numChunks, rand.New(fountain.NewMersenneTwister(200)), solitonDistribution(numChunks))

	var msg = []byte(str) // copy of str, as EncodeLTBlock is destructive to msg
	idsToEncode := ids(int(float64(numChunks) * e.redundancyFactor))
	lubyBlocks := fountain.EncodeLTBlocks(msg, idsToEncode, codec)

	ret := make([]string, len(lubyBlocks))
	for i, block := range lubyBlocks {
		ret[i] = e.frame(block.BlockCode, len(str), crc, block.Data)
	}
	return ret, nil
}

// SetRedundancyFactor changes the value of redundancy factor.
func (e *Encoder) SetRedundancyFactor(rf float64) {
	e.redundancyFactor = rf
}

func (e *Encoder) frame(blockCode int64, total int, crc uint32, data []byte) string {
	return fmt.Sprintf("%d/%d/%d/%08x|%s", blockCode, e.chunkLen, total, crc, string(data))
}

func numberOfChunks(length, chunkLen int) int {
	n := length / chunkLen
	if length%chunkLen > 0 {
		n++
	}
	return n
}

// PayloadCRC returns the CRC-32 (IEEE) of the payload bytes.
func PayloadCRC(data []byte) uint32 {
	return crc32.ChecksumIEEE(data)
}
