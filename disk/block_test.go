package disk

import (
	"bytes"
	"encoding/binary"
	"os"
	"path"
	"testing"
)

const (
	tmpTestFile = "cfs_disk_test"
)

// testFilePattern is used to fill the test file, having a repeat pattern helps
// on testing writeBlock. We can check the data before/after the block we
// wrote in is not collapsed
var testFilePattern = []byte{'a', 'b', 'c', 'd', 'e', 'f', 'g'}

func setUpBlockTestFile(length int, t *testing.T) *os.File {
	f, err := os.OpenFile(
		path.Join(os.TempDir(), tmpTestFile),
		os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0600)
	if err != nil {
		t.Fatalf("open tmp test file got error: %v", err)
		return nil
	}

	for length > 0 {
		if length < len(testFilePattern) {
			_, err = f.Write(testFilePattern[:length])
		} else {
			_, err = f.Write(testFilePattern)
		}
		if err != nil {
			t.Fatalf("write tmp test file got error: %v", err)
		}
		length = length - len(testFilePattern)
	}
	return f
}

func TestPartlyReadBlock(t *testing.T) {
	defer os.Remove(tmpTestFile)
	f := setUpBlockTestFile(payloadSize, t)

	// writeBlock to set a correct CRC
	b := newBlock()
	err := writeBlock(f, b, 4)
	if err != nil {
		t.Errorf("error = %v", err)
	}
	// try to read out a block
	rb := newBlock()
	err = readBlock(f, rb, 4)

	// FIXME do we expect error = io.EOF here?
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if rb.right != payloadSize {
		t.Fatalf("expect get data length = %v got %v", payloadSize, rb.right)
	}
}

func BenchmarkBlockWrite(b *testing.B) {
	for i := 0; i < b.N; i++ {
		f, _ := os.OpenFile(
			path.Join(os.TempDir(), tmpTestFile),
			os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0600)
		writeBlock(f, newBlock(), i)
		defer os.Remove(tmpTestFile)
	}
}

func TestCRCErrCheck(t *testing.T) {
	tests := []struct {
		fileSize int
		index    int
	}{
		// wrong crc
		{blockSize * 2, 1},
		// small file (no crc)
		{2, 0},
		// CRC error since it is appending to the last piece of the block
		{blockSize * 10, 0},

	}
	defer os.Remove(tmpTestFile)
	for i, tt := range tests {
		f := setUpBlockTestFile(tt.fileSize, t)
		// writeBlock to set a correct CRC

		b := newBlock()
		err := readBlock(f, b, tt.index)
		if err != ErrBadCRC {
			t.Errorf("%d: expect error %v got %v", i, ErrBadCRC, err)
		}
	}
}

func TestReadWriteBlock(t *testing.T) {
	tests := []struct {
		fileSize int
		index    int
		writeLen int
	}{
		// basic
		{blockSize, 0, payloadSize},
		// none-zero index
		{blockSize * 5, 2, payloadSize},
		// write partly block
		{blockSize * 5 - 10, 4, payloadSize},
		// empty data
		{0, 0, 0},
		// empty data with just fit block size
		{blockSize * 25,0, 0},
	}

	defer os.Remove(tmpTestFile)
	for i, tt := range tests {
		f := setUpBlockTestFile(tt.fileSize, t)

		block := newBlock()
		for i := 0; i < payloadSize; i++ {
			block.buf[i] = 'X'
		}
		err := writeBlock(f, block, tt.index)
		if err != nil {
			t.Errorf("%d: error = %v", i, err)
		}

		rb := newBlock()
		err = readBlock(f, rb, tt.index)
		if err != nil {
			t.Fatalf("%d: error = %v", i, err)
		}
		if rb.right != payloadSize {
			t.Fatalf("%d: expect read length %v got %v", i, payloadSize, rb.right)
		}
		if !bytes.Equal(rb.buf, block.buf) {
			t.Errorf("%d: expect block data %v got %v", i, rb.buf[:100], block.buf[:100])
		}

		// check crc
		b := make([]byte, crc32Len)
		if err := seekToIndex(f, tt.index); err != nil {
			t.Errorf("%d: error = %v", i, err)
		}
		if _, err := f.Read(b); err != nil {
			t.Errorf("%d: error = %v", i, err)
		}
		crc := binary.BigEndian.Uint32(b)
		if crc != block.GetCRC() {
			t.Errorf("%d: expect crc %v got %v",
				i, crc, block.GetCRC())
		}
	}
}
