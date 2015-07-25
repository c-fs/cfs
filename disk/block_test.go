package disk

import (
	"bytes"
	"encoding/binary"
	// "errors"
	// "hash/crc32"
	// "strings"
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
	// 90 bytes file, block size = 20 bytes
	f := setUpBlockTestFile(90, t)

	bm := newBlockManager(20, f)

	// writeBlock to set a correct CRC
	b := &Block{make([]byte, 6), 0, 6, true}
	err := bm.writeBlock(b, 4)
	if err != nil {
		t.Errorf("error = %v", err)
	}
	// try to read out a block
	rb := &Block{make([]byte, 16), 0, 16, true}
	err = bm.readBlock(rb, 4)

	// FIXME do we expect error = io.EOF here?
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if rb.size != 6 {
		t.Fatalf("expect get data length = %v got %v", 6, rb.size)
	}
}

func TestCRCErrCheck(t *testing.T) {
	tests := []struct {
		fileSize int
		bs       int
		index    int
	}{
		// wrong crc
		{20, 10, 1},
		// small file (no crc)
		{2, 10, 0},
		// CRC error since it is appending to the last piece of the block
		{100, 10, 0},

	}
	defer os.Remove(tmpTestFile)
	for i, tt := range tests {
		f := setUpBlockTestFile(tt.fileSize, t)
		bm := newBlockManager(tt.bs, f)

		// writeBlock to set a correct CRC
		b := &Block{make([]byte, 6), 0, 6, true}
		err := bm.readBlock(b, tt.index)
		if err != ErrBadCRC {
			t.Errorf("%d: expect error %v got %v", i, ErrBadCRC, err)
		}
	}
}

func TestReadWriteBlock(t *testing.T) {
	tests := []struct {
		fileSize int
		bs       int
		index    int
		writeLen int
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
	// errtests := []struct {
	// 	fileSize int
	// 	bs       int
	// 	index    int
	// 	writeLen int
	// 	err      error
	// }{
	// 	// bs < len(data + crc)
	// 	{20, 20, 0, 40, ErrPayloadSizeTooLarge},
	// 	// too small block size
	// 	{20, 1, 0, 0, ErrPayloadSizeTooLarge},
	// 	// negative index
	// 	{20, 20, -1, 16, errors.New("invalid argument")},
	// }

	defer os.Remove(tmpTestFile)
	for i, tt := range tests {
		f := setUpBlockTestFile(tt.fileSize, t)

		block := newBlock(tt.writeLen)
		for i := 0; i < tt.writeLen; i++ {
			block.buf[i] = 'X'
		}
		bm := newBlockManager(tt.bs - crc32Len, f)
		err := bm.writeBlock(block, tt.index)
		if err != nil {
			t.Errorf("%d: error = %v", i, err)
		}

		readBlock := newBlock(tt.writeLen)
		err = bm.readBlock(readBlock, tt.index)
		if err != nil {
			t.Errorf("%d: error = %v", i, err)
		}
		if block.size != tt.writeLen {
			t.Errorf("%d: expect read length %v got %v", i, tt.writeLen, block.size)
		}
		if !bytes.Equal(block.buf, readBlock.buf) {
			t.Errorf("%d: expect block data %v got %v", i, block.buf, readBlock.buf)
		}

		// check crc
		b := make([]byte, crc32Len)
		if err := bm.seekToIndex(tt.index); err != nil {
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

		// check other bytes in file are not collapsed
		prevBlockLastByte := tt.index*tt.bs - 1
		nextBlockFirstByte := (tt.index + 1) * tt.bs
		b = make([]byte, 1)
		if prevBlockLastByte >= 0 {
			if _, err := f.Seek(int64(tt.index*tt.bs-1), os.SEEK_SET); err != nil {
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
			if err := bm.seekToIndex(tt.index+1); err != nil {
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

	// // error test
	// for i, tt := range errtests {
	// 	f := setUpBlockTestFile(tt.fileSize, t)
	// 	p := make([]byte, tt.writeLen)
	// 	for i := int64(0); i < tt.writeLen; i++ {
	// 		p[i] = 'X'
	// 	}
	// 	err := writeBlock(f, tt.index, tt.bs, p)
	// 	if !strings.Contains(err.Error(), tt.err.Error()) {
	// 		t.Errorf("error test %d: expect error %v got %v",
	// 			i, tt.err, err)
	// 	}

	// 	_, err = readBlock(f, p, tt.index, tt.bs)
	// 	if !strings.Contains(err.Error(), tt.err.Error()) {
	// 		t.Errorf("error test %d: expect error %v got %v",
	// 			i, tt.err, err)
	// 	}

	// 	if err = f.Close(); nil != err {
	// 		t.Fatal("can not close test file: ", err)
	// 	}
	// }
}
