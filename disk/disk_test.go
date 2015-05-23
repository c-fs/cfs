package disk

import (
	"os"
	"path"
	"testing"
)

func fillPattern(buf []byte, fillLen int64) {
	for i := int64(0); i < fillLen; i++ {
		buf[i] = testFilePattern[i%int64(len(testFilePattern))]
	}
}

func setUpDiskTestFile(length int64, blockSize int64, t *testing.T) string {
	fileName := path.Join(os.TempDir(), tmpTestFile)
	f, err := os.OpenFile(fileName, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0600)
	if err != nil {
		t.Fatalf("can not open test file: %v", err)
		return ""
	}
	defer f.Close()

	payloadSize := blockSize - crc32Len
	b := make([]byte, payloadSize)
	index := int64(0)
	fillLen := payloadSize
	fillPattern(b, fillLen)
	// fill tmp test file with test pattern
	for length > 0 {
		if length < payloadSize {
			fillLen = length
		}
		if err := writeBlock(f, index, blockSize, b[:fillLen]); err != nil {
			t.Fatalf("can not write to test file: %v", err)
		}
		length = length - payloadSize
		index++
	}

	if err = f.Sync(); nil != err {
		t.Fatalf("can not sync test file: %v", err)
	}
	return fileName
}

func TestReadBlock(t *testing.T) {

}

func TestWriteBlock(t *testing.T) {
	tests := []struct {
		fileSize  int64
		blockSize int64
		offSet    int64
		writeLen  int64
	}{
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
	}

	defer os.Remove(tmpTestFile)
	for i, tt := range tests {
		filePath := setUpDiskTestFile(tt.fileSize, tt.blockSize, t)
		// write
		p := make([]byte, tt.writeLen)
		for i := int64(0); i < tt.writeLen; i++ {
			p[i] = 'X'
		}
		_, err := WriteAt(filePath, p, tt.offSet)
		if err != nil {
			t.Errorf("%d: error = %v", i, err)
		}
	}

}
