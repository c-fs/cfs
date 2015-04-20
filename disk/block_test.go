package disk

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path"
	"strings"
	"testing"
)

const (
	tmpTestFile = "cfs_disk_block_test"
)

// testFilePattern is used to fill the test file, having a repeat pattern helps
// on testing writeBlock. We can check the data before/after the block we
// wrote in is not collapsed
var testFilePattern = []byte{'a', 'b', 'c', 'd', 'e', 'f', 'g'}

func setUpTestFile(length int64, t *testing.T) *os.File {
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
	f := setUpTestFile(90, t)

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
	if err != io.EOF {
		t.Errorf("expect error %v got %v", io.EOF, err)
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
		{100, 20, 0, 0},
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
		// negative index
		{20, 20, -1, 16, errors.New("invalid argument")},
	}

	defer os.Remove(tmpTestFile)
	for i, tt := range tests {
		f := setUpTestFile(tt.fileSize, t)
		p := make([]byte, tt.writeLen)
		for i := int64(0); i < tt.writeLen; i++ {
			p[i] = 'X'
		}

		// write
		err := writeBlock(f, tt.index, tt.bs, p)
		if err != nil {
			t.Errorf("%d: error = %v", i, err)
		}

		rp := make([]byte, tt.writeLen)
		// read
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

		// TODO check crc
		// TODO check other bytes in file are not collapsed

		if err = f.Close(); nil != err {
			t.Fatal("can not close test file: ", err)
		}
	}

	for i, tt := range errtests {
		f := setUpTestFile(tt.fileSize, t)
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
