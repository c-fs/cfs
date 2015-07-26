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


type Block struct{
	buf []byte
	offset int
	size int
}

func (b *Block) GetPayload() []byte {
	return b.buf[b.offset:b.size]
}

func (b *Block) Seek(offset int) {
	b.offset = offset
}

func (b *Block) Copy(offset int, from []byte) int {
	copied := copy(b.buf[offset:], from)
	b.size = max(b.size, offset+copied)
	return copied
}

func (b *Block) GetCRC() uint32 {
	return crc32.Checksum(b.buf[:b.size], crc32cTable)
}

func (b *Block) IsPartial() bool {
	return !(b.offset == 0 && b.size == payloadSize)
}

func (b *Block) Merge(toMerge *Block) {
	b.offset = min(b.offset, toMerge.offset)
	b.size = min(b.size, toMerge.size)
	copy(b.buf[toMerge.offset:toMerge.size], toMerge.GetPayload())
}

func (b *Block) Reset() {
	b.size = 0
	b.offset = 0
}

func newBlock() *Block {
	return &Block{make([]byte, payloadSize), 0, payloadSize}
}

func seekToIndex(s io.Seeker, index int) error {
	_, err := s.Seek(int64(index * blockSize), os.SEEK_SET)
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
	crcBuf := make([]byte, crc32Len)
	if err := seekToIndex(f, index); err != nil {
		return err
	}
	n, err := f.Read(crcBuf)
	// Cannot read full crc
	if n > 0 && n < 4 {
		return ErrBadCRC
	}
	if err != nil {
		return err
	}
	crc := binary.BigEndian.Uint32(crcBuf)
	b.size, err = f.Read(b.buf)
	if err != nil {
		return err
	}
	// Invalid crc
	if crc != b.GetCRC() {
		return ErrBadCRC
	}
	return nil
}

func writeBlock(f io.WriteSeeker, b *Block, index int) error {
	if b.size > payloadSize {
		return ErrPayloadSizeTooLarge
	}
	if err := seekToIndex(f, index); err != nil {
		return err
	}

	crcBuf := make([]byte, crc32Len)
	binary.BigEndian.PutUint32(crcBuf, b.GetCRC())
	_, err := f.Write(crcBuf)
	if err != nil {
		return err
	}
	if err := seekToOffset(f, b.offset); err != nil {
		return err
	}
	_, err = f.Write(b.GetPayload())
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
