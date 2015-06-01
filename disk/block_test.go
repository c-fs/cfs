package disk

import (
	"bytes"
	"encoding/binary"
	"errors"
	"hash/crc32"
	"os"
	"path"
	"strings"
	"testing"
)

const (
	tmpTestFile = "cfs_disk_test"
)

// testFilePattern is used to fill the test file, having a repeat pattern helps
// on testing writeBlock. We can check the data before/after the block we
// wrote in is not collapsed
var testFilePattern = []byte{'a', 'b', 'c', 'd', 'e', 'f', 'g'}

func setUpBlockTestFile(length int64, t *testing.T) *os.File {
	f, err := os.OpenFile(
		path.Join(os.TempDir(), tmpTestFile),
		os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0600)
	if err != nil {
		t.Fatalf("can not open test file: %v", err)
		return nil
	}

	// fill tmp test file with test pattern
	for length > 0 {
		if length < int64(len(testFilePattern)) {
			_, err = f.Write(testFilePattern[0 : length-1])
		} else {
			_, err = f.Write(testFilePattern)
		}
		if err != nil {
			t.Fatalf("can not write to test file: %v", err)
		}
		length = length - int64(len(testFilePattern))
	}

	if err = f.Sync(); nil != err {
		t.Fatalf("can not sync test file: %v", err)
	}
	return f
}

func TestPartlyReadBlock(t *testing.T) {
	defer os.Remove(tmpTestFile)
	// 90 bytes file, block size = 20 bytes
	f := setUpBlockTestFile(90, t)

	// writeBlock to set a correct CRC
	p := make([]byte, 6)
	err := writeBlock(f, 4, 20, p)
	if err != nil {
		t.Errorf("error = %v", err)
	}
	// try to read out a block
	rp := make([]byte, 16)
	n, err := readBlock(f, rp, 4, 20)
	if n != 6 {
		t.Errorf("expect get data length = %v got %v", 6, n)
	}
	// FIXME do we expect error = io.EOF here?
	if err != nil {
		t.Errorf("error = %v", err)
	}
}

func TestCRCErrCheck(t *testing.T) {
	tests := []struct {
		fileSize int64
		bs       int64
		index    int64
	}{
		// wrong crc
		{20, 10, 1},
		// small file (no crc)
		{2, 10, 0},
	}
	defer os.Remove(tmpTestFile)
	p := make([]byte, 6)
	for i, tt := range tests {
		f := setUpBlockTestFile(tt.fileSize, t)
		_, err := readBlock(f, p, tt.index, tt.bs)
		if err != ErrBadCRC {
			t.Errorf("%d: expect error %v got %v", i, ErrBadCRC, err)
		}
	}
}

func TestReadWriteBlock(t *testing.T) {
	tests := []struct {
		fileSize int64
		bs       int64
		index    int64
		writeLen int64
	}{
		// basic
		{20, 20, 0, 16},
		// none-zero index
		{100, 20, 2, 16},
		// write partly block
		{90, 20, 4, 16},
		// empty data
		{0, 20, 0, 0},
		// empty data with just fit block size
		{100, 4, 0, 0},
	}
	errtests := []struct {
		fileSize int64
		bs       int64
		index    int64
		writeLen int64
		err      error
	}{
		// bs < len(data + crc)
		{20, 20, 0, 40, ErrPayloadSizeTooLarge},
		// too small block size
		{20, 1, 0, 0, ErrPayloadSizeTooLarge},
		// CRC error since it is appending to the last piece of the block
		{100, 20, 0, 0, ErrBadCRC},
		// negative index
		{20, 20, -1, 16, errors.New("invalid argument")},
	}

	defer os.Remove(tmpTestFile)
	for i, tt := range tests {
		f := setUpBlockTestFile(tt.fileSize, t)

		// write
		p := make([]byte, tt.writeLen)
		for i := int64(0); i < tt.writeLen; i++ {
			p[i] = 'X'
		}
		err := writeBlock(f, tt.index, tt.bs, p)
		if err != nil {
			t.Errorf("%d: error = %v", i, err)
		}

		// read
		rp := make([]byte, tt.writeLen)
		n, err := readBlock(f, rp, tt.index, tt.bs)
		if err != nil {
			t.Errorf("%d: error = %v", i, err)
		}
		if n != tt.writeLen {
			t.Errorf("%d: expect read length %v got %v", i, tt.writeLen, n)
		}
		if !bytes.Equal(p, rp) {
			t.Errorf("%d: expect block data %v got %v", i, p, rp)
		}

		// check crc
		b := make([]byte, crc32Len)
		if _, err := f.Seek(tt.index*tt.bs, os.SEEK_SET); err != nil {
			t.Errorf("%d: error = %v", i, err)
		}
		if _, err := f.Read(b); err != nil {
			t.Errorf("%d: error = %v", i, err)
		}
		crc := binary.BigEndian.Uint32(b)
		if crc != crc32.Checksum(rp, crc32cTable) {
			t.Errorf("%d: expect crc %v got %v",
				i, crc, crc32.Checksum(rp, crc32cTable))
		}

		// check other bytes in file are not collapsed
		prevBlockLastByte := tt.index*tt.bs - 1
		nextBlockFirstByte := (tt.index + 1) * tt.bs
		b = make([]byte, 1)
		if prevBlockLastByte >= 0 {
			if _, err := f.Seek(tt.index*tt.bs-1, os.SEEK_SET); err != nil {
				t.Errorf("%d: error = %v", i, err)
			}
			if _, err := f.Read(b); err != nil {
				t.Errorf("%d: error = %v", i, err)
			}
			patternIndex := int(prevBlockLastByte) % len(testFilePattern)
			if b[0] != testFilePattern[patternIndex] {
				t.Errorf("%d: expect byte %v got %v",
					i, b[0], testFilePattern[patternIndex])
			}
		}
		if nextBlockFirstByte < tt.fileSize {
			_, err = f.Seek((tt.index+1)*tt.bs, os.SEEK_SET)
			if err != nil {
				t.Errorf("%d: error = %v", i, err)
			}
			_, err = f.Read(b)
			if err != nil {
				t.Errorf("%d: error = %v", i, err)
			}
			patternIndex := int(nextBlockFirstByte) % len(testFilePattern)
			if b[0] != testFilePattern[patternIndex] {
				t.Errorf("%d: expect byte %v got %v",
					i, b[0], testFilePattern[patternIndex])
			}
		}

		// close file
		if err = f.Close(); nil != err {
			t.Fatal("can not close test file: ", err)
		}
	}

	// error test
	for i, tt := range errtests {
		f := setUpBlockTestFile(tt.fileSize, t)
		p := make([]byte, tt.writeLen)
		for i := int64(0); i < tt.writeLen; i++ {
			p[i] = 'X'
		}
		err := writeBlock(f, tt.index, tt.bs, p)
		if !strings.Contains(err.Error(), tt.err.Error()) {
			t.Errorf("error test %d: expect error %v got %v",
				i, tt.err, err)
		}

		_, err = readBlock(f, p, tt.index, tt.bs)
		if !strings.Contains(err.Error(), tt.err.Error()) {
			t.Errorf("error test %d: expect error %v got %v",
				i, tt.err, err)
		}

		if err = f.Close(); nil != err {
			t.Fatal("can not close test file: ", err)
		}
	}
}
