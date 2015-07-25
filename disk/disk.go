package disk

import (
	"io"
	"os"
	"path"
//	"github.com/qiniu/log"
)

// TODO: interface?
type Disk struct {
	// Name is the name of the disk.
	Name string
	// Root is the root path of the disk
	// Usually it is the mount point of a disk or a directory
	// under the mount point.
	Root string
	// PayloadSize
	payloadSize int
}

type BlockReaderStream struct {
	blockIndex int
	blockOffset int
	blockManager *BlockManager
	block *Block
}

func (brs *BlockReaderStream) NextBlock() (*Block, error) {
	brs.block.Reset()
	err := brs.blockManager.readBlock(brs.block, brs.blockIndex)
	brs.block.Seek(brs.blockOffset)
	brs.blockOffset = 0
	brs.blockIndex += 1
	return brs.block, err
}

func NewBlockReaderStream(index, offset int, bm *BlockManager) *BlockReaderStream {
	return &BlockReaderStream{index, offset, bm, bm.newBlock()}

}

func (d *Disk) NewBlockManager(f io.ReadWriteSeeker) *BlockManager {
	return newBlockManager(d.payloadSize, f)
}

// ReadAt reads up to len(p) bytes starting at byte offset off
// from the File into p.
// It returns the number of bytes read and an error, if any.
func (d *Disk) ReadAt(name string, p []byte, off int64) (int, error) {
	dataOffset := int(off)
	name = path.Join(d.Root, name)

	// nil or zero length payload
	if len(p) == 0 {
		return 0, nil
	}

	f, err := os.OpenFile(name, os.O_RDONLY, 0600)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	bm := d.NewBlockManager(f)

	index, offset := bm.getBlockIndexAndOffset(dataOffset)

	stream := NewBlockReaderStream(index, offset, bm)

	read := 0
	for {
		block, err := stream.NextBlock()

		copied := copy(p, block.GetPayload())
		// We just copied some data into p, shrink p
		p = p[copied:]
		// We want to exit the loop for the following cases:
		// 1. There is an error reading block
		// 2. We can't copy a full block into p
		if err != nil || block.size < d.payloadSize || len(p) == 0 {
			return read + copied, err
		}
		read += copied
	}
}

type BlockWriterStream struct {
	blockOffset int
	data []byte
	block *Block
	blockManager *BlockManager
	payloadSize int
	written int
}

// NextBlock gets the next block from the input stream
func (bws *BlockWriterStream) NextBlock(index int) (*Block, error) {
	bws.block.Reset()

	// full block
	if bws.blockOffset == 0 && len(bws.data) >= bws.payloadSize {
		bws.block.Copy(0, bws.data[:bws.payloadSize])
		bws.data = bws.data[bws.payloadSize:]
		bws.written += bws.payloadSize
		return bws.block, nil
	}
	// partial block, needs to merge with existing block

	// read existing block
	err := bws.blockManager.readBlock(bws.block, index)
	if err == io.EOF {
		err = nil
	}
	if err != nil {
		return bws.block, err
	}

	// copy data into it
	payloadLen := bws.payloadSize - bws.blockOffset
	if len(bws.data) > payloadLen {
		bws.written += bws.block.Copy(bws.blockOffset,
			bws.data[:payloadLen])
		bws.data = bws.data[payloadLen:]
	} else {
		bws.written += bws.block.Copy(bws.blockOffset, bws.data)
		bws.data = nil
	}
	bws.blockOffset = 0
	return bws.block, nil
}

func (d *Disk) getDataLength(f *os.File) int {
	fi, err := f.Stat()
	if err != nil {
		return 0
	}
	s := int(fi.Size())
	bm := d.NewBlockManager(f)
	blockNum := (s + bm.blockSize - 1) / bm.blockSize
	return s - blockNum*crc32Len
}


// WriteAt writes len(p) bytes to the File starting at byte offset off.
// It returns the number of bytes written and an error, if any. WriteAt
// returns a non-nil error when n != len(p).
func (d *Disk) WriteAt(name string, p []byte, off int64) (int, error) {
	dataOffset := int(off)
	name = path.Join(d.Root, name)
	// nil or zero length payload
	if len(p) == 0 {
		return 0, nil
	}
	f, err := os.OpenFile(name, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	fileDataLength := d.getDataLength(f)
	bm := d.NewBlockManager(f)

	index, offset := bm.getBlockIndexAndOffset(dataOffset)
	fileDataIndex, _ := bm.getBlockIndexAndOffset(fileDataLength)
	currentIndex := fileDataIndex
	// fill with 0 padding
	for currentIndex < index {
		block := bm.newBlock()
		if currentIndex == fileDataIndex {
			err := bm.readBlock(block, currentIndex)
			if err == io.EOF {
				err = nil
			}
			if err != nil {
				return 0, err
			}
			block.size = d.payloadSize
		}
		bm.writeBlock(block, currentIndex)
		currentIndex += 1
	}

	stream := BlockWriterStream{offset, p, bm.newBlock(), bm,
		bm.payloadSize, 0}

	for {
		block, err := stream.NextBlock(index)

		if err != nil {
			return stream.written, err
		}
		if block.size == 0 {
			break
		}

		err = bm.writeBlock(block, index)
		if err != nil {
			return stream.written, err
		}
		index++;
	}
	return stream.written, nil
}

func (d *Disk) Rename(oldname, newname string) error {
	oldname = path.Join(d.Root, oldname)
	newname = path.Join(d.Root, newname)
	return os.Rename(oldname, newname)
}

func (d *Disk) Remove(name string, all bool) error {
	name = path.Join(d.Root, name)
	if !all {
		return os.Remove(name)
	}
	return os.RemoveAll(name)
}

func (d *Disk) ReadDir(name string) ([]os.FileInfo, error) {
	name = path.Join(d.Root, name)
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return f.Readdir(0)
}

func (d *Disk) Mkdir(name string, all bool) error {
	name = path.Join(d.Root, name)
	if !all {
		return os.Mkdir(name, 0700)
	}
	return os.MkdirAll(name, 0700)
}
