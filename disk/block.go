package disk

import (
	"encoding/binary"
	"errors"
	"hash/crc32"
	"io"
	"os"
)

var (
	crc32cTable = crc32.MakeTable(crc32.Castagnoli)
	crc32Len    = 4
	blockSize   = 4096
	payloadSize = blockSize - crc32Len

	// ErrPayloadSizeTooLarge indicates the input payload size is too big
	ErrPayloadSizeTooLarge = errors.New("disk: bad payload size")
	// ErrBadCRC indicates there is not CRC can be found in the block
	ErrBadCRC = errors.New("disk: not a valid CRC")
)

// Block is a buffer aligned with disk data block with two offset (left and right)
// representing the start and end of the effective payload inside the buffer
type Block struct {
	buf   []byte
	left  int
	right int
}

// IsEmpty tells if the buffer's effective payload is empty
func (b *Block) IsEmpty() bool {
	return b.left == b.right
}

// Payload returns the effective payload as a byte array
func (b *Block) Payload() []byte {
	return b.buf[b.left:b.right]
}

// Copy copies a byte array to the offset of the buffer
// It also updates the effective payload offsets
func (b *Block) Copy(offset int, from []byte) int {
	copied := copy(b.buf[offset:], from)
	b.StartFrom(min(b.left, offset))
	b.EndAt(max(b.right, offset+copied))
	return copied
}

// StartFrom sets the block's left offset
func (b *Block) StartFrom(offset int) {
	b.left = offset
}

// EndAt sets the block's right offset
func (b *Block) EndAt(offset int) {
	b.right = offset
}

// CRC calculates the crc of the effective payload
func (b *Block) CRC() uint32 {
	return crc32.Checksum(b.buf[b.left:b.right], crc32cTable)
}

// IsPartial checks if effective payload is a partial payload
func (b *Block) IsPartial() bool {
	return !(b.left == 0 && b.right == payloadSize)
}

// Merge uses current block as a base and
// merges another block on top of it
func (b *Block) Merge(toMerge *Block) {
	b.StartFrom(min(b.left, toMerge.left))
	b.EndAt(max(b.right, toMerge.right))
	copy(b.buf[toMerge.left:toMerge.right], toMerge.Payload())
}

// Reset resets the effective payload to 0
func (b *Block) Reset() {
	b.StartFrom(0)
	b.EndAt(0)
}

func newBlock() *Block {
	return &Block{make([]byte, payloadSize), 0, payloadSize}
}

func seekToIndex(s io.Seeker, index int) error {
	_, err := s.Seek(int64(index*blockSize), os.SEEK_SET)
	return err
}

func seekToOffset(s io.Seeker, offset int) error {
	_, err := s.Seek(int64(offset), os.SEEK_CUR)
	return err
}

func blockIndexAndOffset(dataOffset int) (int, int) {
	index := dataOffset / payloadSize
	offset := dataOffset - index*payloadSize
	return index, offset
}

func readBlock(f io.ReadSeeker, b *Block, index int) error {
	b.Reset()
	if err := seekToIndex(f, index); err != nil {
		return err
	}
	buf := make([]byte, blockSize)
	n, err := f.Read(buf)
	// Cannot read full crc
	if n > 0 && n < 4 {
		return ErrBadCRC
	}
	if err != nil {
		return err
	}
	crc := binary.BigEndian.Uint32(buf[:crc32Len])
	v := copy(b.buf, buf[crc32Len:n])
	b.EndAt(v)
	// Invalid crc
	if crc != b.CRC() {
		return ErrBadCRC
	}
	return nil
}

func writeBlock(f io.WriteSeeker, b *Block, index int) error {
	if b.right > payloadSize {
		return ErrPayloadSizeTooLarge
	}
	if b.left != 0 {
		panic("block is not left aligned")
	}
	if err := seekToIndex(f, index); err != nil {
		return err
	}

	crcBuf := make([]byte, crc32Len+len(b.Payload()))
	binary.BigEndian.PutUint32(crcBuf, b.CRC())
	copy(crcBuf[crc32Len:], b.Payload())
	_, err := f.Write(crcBuf)
	if err != nil {
		return err
	}
	return nil
}

func min(a, b int) int {
	if a > b {
		return b
	}
	return a
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
