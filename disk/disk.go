package disk

import (
	"io"
	"log"
	"os"
)

// WriteAt writes len(p) bytes from p to the file at path at offset off.
// It returns the number of bytes written from p (0 <= n <= len(p)) and
// any error encountered that caused the write to stop early.
// WriteAt must return a non-nil error if it returns n < len(p).
// TODO(xiang90): []byte or io.reader?
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

	// offset index that starts to write in the next time
	i := min(off, flen)
	// number of bytes that has written
	n := 0
	// index that ends writing
	end := off + int64(len(p))
	rbuf := make([]byte, psize)
	for i < end {
		pidx := i / psize
		poff := i - pidx*psize
		wbuf := make([]byte, 0)

		// read out the head of data in block
		if poff > 0 {
			rn, err := readBlock(f, rbuf, pidx, bsize)
			if err != nil && err != io.EOF {
				return n, err
			}
			if poff < rn {
				log.Panicf("unexpected insufficient read")
			}
			wbuf = append(wbuf, rbuf[:poff]...)
		}

		left := end - i
		pleft := psize - poff
		wn := min(left, pleft)
		var wrtdata []byte
		if i+wn < off {
			wrtdata = make([]byte, wn)
		} else if i > off {
			wrtdata = p[n : n+int(wn)]
		} else {
			wrtdata = append(make([]byte, off-i), p[:i+wn-off]...)
		}
		wbuf = append(wbuf, wrtdata...)

		// read out the tail of data in block
		if poff+wn < psize && end < flen {
			rn, err := readBlock(f, rbuf, pidx, bsize)
			if err != nil && err != io.EOF {
				return n, err
			}
			if poff+wn < rn {
				log.Panicf("unexpected insufficient read")
			}
			wbuf = append(wbuf, p[poff+wn:]...)
		}

		err := writeBlock(f, pidx, bsize, wbuf)
		if err != nil {
			return n, err
		}
		n += int(wn)
		i += int64(len(wbuf))
	}
	return n, nil
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
