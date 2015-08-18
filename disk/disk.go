package disk

import (
	"io"
	"os"
	"path"
)

// TODO: interface?
type Disk struct {
	// Name is the name of the disk.
	Name string
	// Root is the root path of the disk
	// Usually it is the mount point of a disk or a directory
	// under the mount point.
	Root string
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

	index, offset := blockIndexAndOffset(dataOffset)
	stream := &BlockReaderStream{index, offset, f}
	read := 0
	for {
		block, err := stream.NextBlock()
		copied := copy(p, block.Payload())
		// We just copied some data into p, shrink p
		p = p[copied:]
		read += copied
		if err != nil {
			return read, err
		}
		if len(p) == 0 {
			return read, nil
		}
		if block.right < payloadSize  {
			return read, io.EOF
		}
	}
}

// getDataSize returns the size of data in file (excluding the crc header size)
func (d *Disk) getDataSize(f *os.File) int {
	fi, err := f.Stat()
	if err != nil {
		return 0
	}
	s := int(fi.Size())
	blockNum := (s + blockSize - 1) / blockSize
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
	fileDataLength := d.getDataSize(f)
	index, offset := blockIndexAndOffset(dataOffset)
	fileDataIndex, _ := blockIndexAndOffset(fileDataLength)
	currentIndex := fileDataIndex
	// fill with 0 padding
	for currentIndex < index {
		block := newBlock()
		if currentIndex == fileDataIndex {
			err := readBlock(f, block, currentIndex)
			if err != nil && err != io.EOF {
				return 0, err
			}
			block.right = payloadSize
		}
		writeBlock(f, block, currentIndex)
		currentIndex += 1
	}

	stream := BlockWriterStream{offset, p, f}
	written := 0
	for {
		block, err := stream.NextBlock()
		if block.IsEmpty() {
			return written, err
		}
		toWrite := len(block.Payload())
		if block.IsPartial() {
			// Merge with existing
			base := newBlock()
			err := readBlock(f, base, index)
			if err != nil && err != io.EOF {
				return written, err
			}
			base.Merge(block)
			block = base
		}
		err = writeBlock(f, block, index)
		if err != nil {
			return written, err
		}

		written += toWrite
		index++;
	}
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
