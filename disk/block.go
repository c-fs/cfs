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

func (b *Block) Copy(from []byte) int {
	copied := copy(b.buf[b.offset:], from)
	b.size = max(b.size, b.offset+copied)
	return copied
}

func (b *Block) GetCRC() uint32 {
	return crc32.Checksum(b.GetPayload(), crc32cTable)
}

func (b *Block) Reset() {
	b.size = 0
	b.offset = 0
}

type BlockManager struct{
	payloadSize int
	blockSize int
	rws io.ReadWriteSeeker
}

func newBlockManager(payloadSize int, rws io.ReadWriteSeeker) *BlockManager {
	return &BlockManager{payloadSize, payloadSize + crc32Len, rws}
}

func newBlock(size int) *Block {
	return &Block{make([]byte, size), 0, size}
}

// newBlock creates an empty block
func (bm *BlockManager) newBlock() *Block {
	return newBlock(bm.payloadSize)
}

func (bm *BlockManager) seekToIndex(index int) error {
	_, err := bm.rws.Seek(int64(index * bm.blockSize), os.SEEK_SET)
	return err
}

func (bm *BlockManager) seekToOffset(offset int) error {
	_, err := bm.rws.Seek(int64(offset), os.SEEK_CUR)
	return err
}

func (bm *BlockManager) getBlockIndexAndOffset(dataOffset int) (int, int) {
	index := dataOffset / bm.payloadSize
	offset := dataOffset - index*bm.payloadSize
	return index, offset
}

func (bm *BlockManager) readBlock(b *Block, index int) error {
	crcBuf := make([]byte, crc32Len)
	if err := bm.seekToIndex(index); err != nil {
		return err
	}
	n, err := bm.rws.Read(crcBuf)
	// Cannot read full crc
	if n > 0 && n < 4 {
		return ErrBadCRC
	}
	if err != nil {
		return err
	}
	crc := binary.BigEndian.Uint32(crcBuf)
	b.size, err = bm.rws.Read(b.buf)
	if err != nil {
		return err
	}
	// Invalid crc
	if crc != b.GetCRC() {
		return ErrBadCRC
	}
	return nil
}

func (bm *BlockManager) writeBlock(b *Block, index int) error {
	if b.size > bm.payloadSize {
		return ErrPayloadSizeTooLarge
	}
	if err := bm.seekToIndex(index); err != nil {
		return err
	}

	crcBuf := make([]byte, crc32Len)
	binary.BigEndian.PutUint32(crcBuf, b.GetCRC())
	_, err := bm.rws.Write(crcBuf)
	if err != nil {
		return err
	}
	if err := bm.seekToOffset(b.offset); err != nil {
		return err
	}
	_, err = bm.rws.Write(b.GetPayload())
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
