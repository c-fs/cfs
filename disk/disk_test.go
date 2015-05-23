package disk

import (
	"bytes"
	"os"
	"path"
	"testing"
)

func fillPattern(buf []byte, fillLen int64) {
	for i := int64(0); i < fillLen; i++ {
		buf[i] = testFilePattern[i%int64(len(testFilePattern))]
	}
}

func setUpDiskTestFile(length int64, blockSize int64, t *testing.T) (string, *os.File) {
	fileRoot := os.TempDir()
	fileName := path.Join(fileRoot, tmpTestFile)
	f, err := os.OpenFile(fileName, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0600)
	if err != nil {
		t.Fatalf("open tmp test file got error: %v", err)
		return "", nil
	}

	payloadSize := blockSize - crc32Len
	b := make([]byte, payloadSize)
	index := int64(0)
	fillLen := payloadSize
	fillPattern(b, fillLen)

	for length > 0 {
		if length < payloadSize {
			fillLen = length
		}
		if err := writeBlock(f, index, blockSize, b[:fillLen]); err != nil {
			t.Fatalf("write tmp test file got error: %v", err)
		}
		length = length - payloadSize
		index++
	}
	return fileRoot, f
}

func TestReadWriteDisk(t *testing.T) {
	tests := []struct {
		dataSize  int64
		blockSize int64
		offSet    int64
		writeLen  int64
	}{
		// zero write
		{0, 4096, 0, 0},
		// empty file write one block
		{0, 4096, 0, 4096 - crc32Len},
		// empty file write two blocks
		{0, 4096, 0, (4096 - crc32Len) * 2},
		// empty file write three and half blocks
		{0, 4096, 0, 4096*3 + 2048},
		// empty file write half block,
		{0, 4096, 0, 2048},
		// empty file add padding
		{0, 4096, 4096, 4096*2 + 2048},
		// one block file write one block
		{4096 - crc32Len, 4096, 4096 - crc32Len, 4096 - crc32Len},
		// one block file write two blocks
		{4096 - crc32Len, 4096, 4096 - crc32Len, (4096 - crc32Len) * 2},
		// one block file write three and half blocks
		{4096 - crc32Len, 4096, 4096 - crc32Len, 4096*3 + 2048},
		// one block file write half block
		{4096 - crc32Len, 4096, 4096 - crc32Len, 2048},
		// one block file add padding
		{4096 - crc32Len, 4096, 4096, 4096*2 + 2048},
		// one block file over write first block
		{4096 - crc32Len, 4096, 2048, 4096*2 + 2048},
		// two and half blocks file write one block
		{4096*2 + 2048, 4096, 4096, 4096 - crc32Len},
		// two and half blocks file wirte half block
		{4096*2 + 2048, 4096, 4096*2 + 2048, 2036},
		// two and half blocks file write with padding
		{4096*2 + 2048, 4096, 4096*2 + 4084, 4096 - crc32Len},
	}

	defer os.Remove(tmpTestFile)
	for i, tt := range tests {
		fileRoot, f := setUpDiskTestFile(tt.dataSize, tt.blockSize, t)
		d := &Disk{Name: "", Root: fileRoot}
		originalDSize := getDataLength(f, tt.blockSize)
		expectedDSize := int64(max(int(originalDSize), int(tt.offSet+tt.writeLen)))

		// write
		p := make([]byte, tt.writeLen)
		for i := int64(0); i < tt.writeLen; i++ {
			p[i] = 'X'
		}
		wn, err := d.WriteAt(tmpTestFile, p, tt.offSet)
		if err != nil {
			t.Errorf("%d: error = %v", i, err)
		}

		// check data size after write
		dSize := getDataLength(f, tt.blockSize)
		if dSize != expectedDSize {
			t.Errorf("%d: expect data length %d, got %d", i, expectedDSize, dSize)
		}

		// read
		r := make([]byte, tt.writeLen)
		rn, err := d.ReadAt(tmpTestFile, r, tt.offSet)
		if err != nil {
			t.Errorf("%d: error = %v", i, err)
		}

		// check write length and read length
		if wn != rn {
			t.Errorf("%d: writen length %d is not equal to read length %d",
				i, wn, rn)
		}

		// check write data and read data
		if !bytes.Equal(p, r) {
			t.Errorf("%d: writen in data %x is not the same as read out data %x",
				i, p, r)
		}
	}
}

func TestReadNonExistFile(t *testing.T) {
	tests := []struct {
		fileRoot  string
		fileName  string
		blockSize int64
		offSet    int64
		writeLen  int64
	}{
		{"nowhere_exist", "no", 4096, 0, 50},
	}
	for i, tt := range tests {
		d := &Disk{Name: "", Root: tt.fileRoot}
		p := make([]byte, tt.writeLen)
		for i := int64(0); i < tt.writeLen; i++ {
			p[i] = 'X'
		}
		_, err := d.ReadAt(tt.fileName, p, tt.offSet)
		if !os.IsNotExist(err) {
			t.Errorf("%d: expect file not exist, got error = %v", i, err)
		}
	}
}
