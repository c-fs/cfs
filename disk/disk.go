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
	// nil or zero length payload
	if len(p) == 0 {
		return 0, nil
	}

	// block size
	bsize := blockSize(path)
	// payload size
	psize := bsize - crc32Len

	f, err := os.OpenFile(path, os.O_RDWR, 0600)
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
