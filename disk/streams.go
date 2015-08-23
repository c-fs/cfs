package disk

import (
	"io"
)

type BlockReaderStream struct {
	blockIndex  int
	blockOffset int
	file        io.ReadSeeker
}

func (brs *BlockReaderStream) NextBlock() (*Block, error) {
	block := newBlock()
	err := readBlock(brs.file, block, brs.blockIndex)
	block.StartFrom(brs.blockOffset)
	brs.blockOffset = 0
	brs.blockIndex += 1
	return block, err
}

type BlockWriterStream struct {
	blockOffset int
	data        []byte
	file        io.ReadWriteSeeker
}

// NextBlock gets the next block from the input stream
func (bws *BlockWriterStream) NextBlock() (*Block, error) {
	block := newBlock()
	block.EndAt(0)
	// if this is a full block, return it
	if bws.blockOffset == 0 && len(bws.data) >= payloadSize {
		block.buf = bws.data[:payloadSize]
		block.EndAt(payloadSize)
		bws.data = bws.data[payloadSize:]
		return block, nil
	}
	// if this is a partial block, copy data into it
	payloadLen := payloadSize - bws.blockOffset
	block.StartFrom(bws.blockOffset)
	if len(bws.data) > payloadLen {
		block.Copy(bws.blockOffset, bws.data[:payloadLen])
		bws.data = bws.data[payloadLen:]
	} else {
		block.Copy(bws.blockOffset, bws.data)
		bws.data = nil
	}
	bws.blockOffset = 0
	return block, nil
}
