package disk

import (
	"io"
	"log"
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
	name = path.Join(d.Root, name)

	// nil or zero length payload
	if len(p) == 0 {
		return 0, nil
	}

	// block size
	bsize := blockSize(name)
	// payload size
	psize := bsize - crc32Len

	f, err := os.OpenFile(name, os.O_RDONLY, 0600)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	stIdx := off / psize
	stOff := off - stIdx*psize
	// how many bytes have been read
	read := 0
	for {
		buf := make([]byte, bsize-crc32Len)
		n, err := readBlock(f, buf, stIdx, bsize)
		// If the read starts with a non-aligend position,
		// only copy partial block
		copied := copy(p, buf[stOff:n])
		stOff = 0
		// We just copied some data into p, shrink p
		p = p[copied:]
		// We want to exit the loop for 3 cases:
		// 1. There is an error reading block
		// 2. We read a partial block -- reach the end of the file
		// 3. We can't copy into p anymore -- p is filled up
		if err != nil || n < bsize-crc32Len || len(p) == 0 {
			return read + copied, err
		}
		read += copied
		stIdx++
	}
}

// WriteAt writes len(p) bytes to the File starting at byte offset off.
// It returns the number of bytes written and an error, if any. WriteAt
// returns a non-nil error when n != len(p).
func (d *Disk) WriteAt(name string, p []byte, off int64) (int, error) {
	name = path.Join(d.Root, name)

	// nil or zero length payload
	if len(p) == 0 {
		return 0, nil
	}

	// block size
	bsize := blockSize(name)
	// payload size
	psize := bsize - crc32Len

	f, err := os.OpenFile(name, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	dlen := getDataLength(f, bsize)
	// start index
	var st int64
	var padding int
	var data []byte
	if off <= dlen {
		st, padding = off, 0
		data = p
	} else {
		st, padding = dlen, int(off-dlen)
		data = append(make([]byte, padding), p...)
	}
	// index that follows the last written byte
	end := st + int64(len(data))

	n := 0
	stIdx, endIdx := st/psize, end/psize
	stOff, endOff := st-stIdx*psize, end-endIdx*psize
	// fast path for writing at one non-full block
	if stIdx == endIdx {
		if err := fillBlock(f, stIdx, stOff, bsize, data); err != nil {
			return 0, err
		}
		n += len(data)
		return max(n-padding, 0), nil
	}
	// head block
	if stOff > 0 {
		if err := fillBlock(f, stIdx, stOff, bsize, data[:psize-stOff]); err != nil {
			return max(n-padding, 0), err
		}
		data = data[psize-stOff:]
		n += int(psize - stOff)
		stIdx++
	}
	// middle blocks
	for i := stIdx; i < endIdx; i++ {
		err := writeBlock(f, i, bsize, data[:psize])
		if err != nil {
			return max(n-padding, 0), err
		}
		data = data[psize:]
		n += int(psize)
	}
	// tail block
	if endOff > 0 {
		if err := fillBlock(f, endIdx, 0, bsize, data); err != nil {
			return max(n-padding, 0), err
		}
		n += len(data)
	}
	return max(n-padding, 0), nil
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

// fillBlock fills the partial block starting from the given offset with the
// given data. It first reads out the block, fills in the given data, and
// writes the block back.
// The given index and offset should locate an existing point in the file.
// The given bsize indicates the size of each block.
// The given data should all be in the target block, and not exceed the
// block boundary.
func fillBlock(f *os.File, index, offset, bsize int64, data []byte) error {
	buf := make([]byte, bsize-crc32Len)
	n, err := readBlock(f, buf, index, bsize)
	if err != nil && err != io.EOF {
		return err
	}
	switch {
	case n < offset:
		log.Panicf("offset should be not bigger than file size")
	case n < offset+int64(len(data)):
		buf = append(buf[:offset], data...)
	default:
		copy(buf[offset:offset+int64(len(data))], data)
		buf = buf[:n]
	}
	return writeBlock(f, index, bsize, buf)
}

// blockSize returns the block size of the file at given path.
func blockSize(name string) int64 {
	// TODO (xiang90): implement it
	return 4096
}

func getDataLength(f *os.File, bsize int64) int64 {
	fi, err := f.Stat()
	if err != nil {
		return 0
	}
	s := fi.Size()
	blockNum := (s + bsize - 1) / bsize
	return s - blockNum*crc32Len
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
