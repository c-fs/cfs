package disk

import (
	"io"
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
		st = off
		data = p
	} else {
		st = flen
		data = append(make([]byte, off-flen), data...)
	}
	// index that ends writing
	end := st + int64(len(data))
	// has trailing data in file after the end position
	hasTrailing := end < flen

	stPidx, stPoff := offToPayloadPos(st, psize)
	endPidx, endPoff := offToPayloadPos(end, psize)
	// read buffer
	rbuf := make([]byte, psize)
	// fast path for writing at one block
	if stPidx == endPidx {
		_, err := readBlock(f, rbuf, stPidx, bsize)
		if err != nil && err != io.EOF {
			return 0, err
		}
		var wbuf []byte
		if hasTrailing {
			copy(rbuf[stPoff:endPoff], data)
			wbuf = rbuf
		} else {
			wbuf = append(rbuf[:stPoff], data...)
		}
		err = writeBlock(f, stPidx, bsize, wbuf)
		if err != nil {
			return 0, err
		}
		return len(p), nil
	}

	// number of bytes that has written
	n := 0
	// head block
	if stPoff > 0 {
		_, err := readBlock(f, rbuf, stPidx, bsize)
		if err != nil && err != io.EOF {
			return n, err
		}
		wbuf := append(rbuf[:stPoff], data[:psize-stPoff]...)
		data = data[psize-stPoff:]
		err = writeBlock(f, stPidx, bsize, wbuf)
		if err != nil {
			return n, err
		}
		stPidx++
		n += int(psize - stPoff)
	}
	// middle blocks
	for i := stPidx; i < endPidx; i++ {
		err := writeBlock(f, stPidx, bsize, data[:psize])
		data = data[psize:]
		if err != nil {
			return n, err
		}
		n += int(psize)
	}
	// tail block
	if endPoff > 0 {
		var wbuf []byte
		if hasTrailing {
			rn, err := readBlock(f, rbuf, endPidx, bsize)
			if err != nil && err != io.EOF {
				return n, err
			}
			copy(rbuf[:endPoff], data[:endPoff])
			wbuf = rbuf[:rn]
		} else {
			wbuf = data[:endPoff]
		}
		err := writeBlock(f, stPidx, bsize, wbuf)
		if err != nil {
			return n, err
		}
		n += int(endPoff)
	}
	return n, nil
}

func offToPayloadPos(off, psize int64) (bidx, boff int64) {
	bidx = off / psize
	boff = off - bidx*psize
	return
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
