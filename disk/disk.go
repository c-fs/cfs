package disk

import (
	"io"
	"log"
	"os"
)

// WriteAt writes len(p) bytes to the File starting at byte offset off.
// It returns the number of bytes written and an error, if any. WriteAt
// returns a non-nil error when n != len(p).
func WriteAt(path string, p []byte, off int64) (int, error) {
	// block size
	bsize := blockSize(path)
	// payload size
	psize := bsize - crc32Len

	f, err := os.OpenFile(path, os.O_WRONLY, 0600)
	if err != nil {
		return 0, err
	}

	flen := getFileLength(f)
	// start index
	var st int64
	var data []byte
	if off <= flen {
		st, data = off, p
	} else {
		st, data = flen, append(make([]byte, off-flen), p...)
	}
	// index that ends writing
	end := st + int64(len(data))

	n := 0
	stIdx, endIdx := st/psize, end/psize
	for i := stIdx; i < endIdx; i++ {
		// length of data to write this time
		var l int64
		// the block offset
		head := off - i*psize
		if l = psize - head; l > int64(len(data)) {
			l = int64(len(data))
		}
		if head == 0 {
			if err := writeBlock(f, i, bsize, data[:l]); err != nil {
				return n, err
			}
		} else {
			if err := fillBlock(f, i, head, bsize, data[:l]); err != nil {
				return n, err
			}
		}
		n += int(l)
		off += l
		data = data[l:]
	}

	return n, nil
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
	}
	return writeBlock(f, index, bsize, buf)
}

// blockSize returns the block size of the file at given path.
func blockSize(path string) int64 {
	// TODO (xiang90): implement it
	return 4096
}

func getFileLength(f *os.File) int64 {
	fi, err := f.Stat()
	if err != nil {
		return 0
	}
	return fi.Size()
}
