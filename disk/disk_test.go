package disk

import (
	"bytes"
	"os"
	"io"
	"path"
	"testing"
)

func fillPattern(buf []byte, fillLen int64) {
	for i := int64(0); i < fillLen; i++ {
		buf[i] = testFilePattern[i%int64(len(testFilePattern))]
	}
}

func setUpDiskTestFile(fileName string, length int64, blockSize int64, t *testing.T) *os.File {
	f, err := os.OpenFile(fileName, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0600)
	if err != nil {
		t.Fatalf("can not open test file: %v", err)
		return nil
	}

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
	return f
}

func newTestDisk(name, root string, mkdir bool) *Disk {
	root = path.Join(os.TempDir(), "cfs", "test", root)
	if mkdir {
		os.MkdirAll(root, 0777)
	}
	return &Disk{Name: name, Root: root}
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

	diskName := "test"
	diskRoot := "wr"
	diskMkdir := true
	diskRemoveAll := true
	d := newTestDisk(diskName, diskRoot, diskMkdir)
	defer d.Remove("", diskRemoveAll)

	for i, tt := range tests {
		f := setUpDiskTestFile(path.Join(d.Root, tmpTestFile), tt.dataSize, tt.blockSize, t)
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
		diskName  string
		fileRoot  string
		fileName  string
		mkdir     bool
		blockSize int64
		offSet    int64
		readLen   int64
	}{
		// read file on non-exist path
		{"disk0", "nowhere_exist", "no", false, 4096, 0, 50},
		// read non-exist file on exist path
		{"disk0", "nowhere_exist", "no", true, 4096, 0, 50},
	}

	for i, tt := range tests {
		d := newTestDisk(tt.diskName, tt.fileRoot, tt.mkdir)

		p := make([]byte, tt.readLen)
		_, err := d.ReadAt(tt.fileName, p, tt.offSet)
		if !os.IsNotExist(err) {
			t.Errorf("%d: expect file not exist, got error = %v", i, err)
		}

		d.Remove("", true)
	}
}

func TestRename(t *testing.T) {
	tests := []struct {
		diskName    string
		fileNameOld string
		fileNameNew string
	}{
		// rename old to new
		{"disk0", "old", "new"},
		// rename to the existing file
		{"disk0", "old", "old"},
	}

	for i, tt := range tests {
		writeLen := int64(50)
		d := newTestDisk(tt.diskName, "rename", true)

		// write
		p := make([]byte, writeLen)
		for i := int64(0); i < writeLen; i++ {
			p[i] = 'X'
		}

		wn, err := d.WriteAt(tt.fileNameOld, p, 0)
		if err != nil {
			t.Errorf("%d: error = %v", i, err)
		}

		err = d.Rename(tt.fileNameOld, tt.fileNameNew)
		if err != nil {
			t.Errorf("%d: error = %v", i, err)
		}

		// read
		r := make([]byte, writeLen)
		rn, err := d.ReadAt(tt.fileNameNew, r, 0)
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

		d.Remove("", true)
	}
}

func TestRenameNonExistFile(t *testing.T) {
	tests := []struct {
		diskName    string
		fileRoot    string
		fileNameOld string
		fileNameNew string
		mkdir       bool
	}{
		// rename file on non-exist path
		{"disk0", "nowhere_exist", "no", "way", false},
		// rename non-exist file on exist path
		{"disk0", "nowhere_exist", "no", "way", true},
	}

	for i, tt := range tests {
		d := newTestDisk(tt.diskName, tt.fileRoot, tt.mkdir)

		err := d.Rename(tt.fileNameOld, tt.fileNameNew)
		if !os.IsNotExist(err) {
			t.Errorf("%d: expect file not exist, got error = %v", i, err)
		}

		d.Remove("", true)
	}
}

func TestRemove(t *testing.T) {
	tests := []struct {
		diskName   string
		files      []string
		removeAll  bool
		fileRemove string
	}{
		// remove one file
		{
			"disk0",
			[]string{"file1", "file2", "file3"},
			false,
			"file2",
		},
		// remove all files
		{
			"disk0",
			[]string{"file1", "file2", "file3"},
			true,
			"",
		},
	}

	for i, tt := range tests {
		writeLen := int64(50)
		d := newTestDisk(tt.diskName, "remove", true)

		// write
		p := make([]byte, writeLen)
		for i := int64(0); i < writeLen; i++ {
			p[i] = 'X'
		}
		for _, f := range tt.files {
			_, err := d.WriteAt(f, p, 0)
			if err != nil {
				t.Errorf("%d: error = %v", i, err)
			}
		}

		// remove
		d.Remove(tt.fileRemove, tt.removeAll)

		// check
		if tt.removeAll {
			for _, f := range tt.files {
				_, err := d.ReadAt(f, p, 0)
				if !os.IsNotExist(err) {
					t.Errorf("%d: expect file not exist, got error = %v", i, err)
				}
			}
		} else {
			_, err := d.ReadAt(tt.fileRemove, p, 0)
			if !os.IsNotExist(err) {
				t.Errorf("%d: expect file not exist, got error = %v", i, err)
			}
		}

		d.Remove("", true)
	}
}

func TestRemoveNonExistFile(t *testing.T) {
	tests := []struct {
		diskName    string
		fileRoot    string
		fileNameOld string
		mkdir       bool
	}{
		// remove file on non-exist path
		{"disk0", "nowhere_exist", "no", false},
		// remove non-exist file on exist path
		{"disk0", "nowhere_exist", "no", true},
	}
	for i, tt := range tests {
		d := newTestDisk(tt.diskName, tt.fileRoot, tt.mkdir)

		err := d.Remove(tt.fileNameOld, false)
		if !os.IsNotExist(err) {
			t.Errorf("%d: expect file not exist, got error = %v", i, err)
		}

		d.Remove("", true)
	}
}

func TestRemoveNonExistPath(t *testing.T) {
	tests := []struct {
		diskName string
		fileRoot string
	}{
		{"disk0", "nowhere_exist"},
	}

	for i, tt := range tests {
		d := newTestDisk(tt.diskName, tt.fileRoot, false)

		err := d.Remove("", true)
		if err != nil {
			t.Errorf("%d: error = %v", i, err)
		}
	}
}

func TestMkDirReadDir(t *testing.T) {
	tests := []struct {
		diskName  string
		fileDir   string
		fileNames []string
		mkdirAll  bool
	}{
		// test one level dir
		{"disk0", "dir", []string{"file1", "file2", "file3"}, false},
		// test multiple levels dir
		{"disk0", "dir/sub/way", []string{"file1", "file2", "file3"}, true},
	}

	for i, tt := range tests {
		writeLen := int64(50)
		d := newTestDisk(tt.diskName, "mkdir", true)

		err := d.Mkdir(tt.fileDir, tt.mkdirAll)
		if err != nil {
			t.Errorf("%d: error = %v", i, err)
		}

		// write
		p := make([]byte, writeLen)
		for i := int64(0); i < writeLen; i++ {
			p[i] = 'X'
		}
		for _, fn := range tt.fileNames {
			_, err := d.WriteAt(path.Join(tt.fileDir, fn), p, 0)
			if err != nil {
				t.Errorf("%d: error = %v", i, err)
			}
		}

		fis, err := d.ReadDir(tt.fileDir)
		if err != nil {
			t.Errorf("%d: error = %v", i, err)
		}

		// check FileInfo
		fExist := make(map[string]string)
		for _, fi := range fis {
			fExist[fi.Name()] = "pending"
		}
		for _, fn := range tt.fileNames {
			if fExist[fn] == "pending" {
				fExist[fn] = "exist"
			} else if fExist[fn] == "exist" {
				t.Errorf("%d: file %s already exist, ReadDir miss FileInfo", i, fn)
			} else if fExist[fn] == "" {
				t.Errorf("%d: file %s not exist, ReadDir miss FileInfo", i, fn)
			}
		}
		d.Remove("", true)
	}
}

func TestReadDirNonExist(t *testing.T) {
	tests := []struct {
		diskName string
		name     string
	}{
		// file
		{"disk0", "nowhere"},
		// dir
		{"disk0", "no/where"},
	}

	for i, tt := range tests {
		d := newTestDisk(tt.diskName, "read-non-exist", true)

		_, err := d.ReadDir(tt.name)
		if !os.IsNotExist(err) {
			t.Errorf("%d: expect file not exist, got error = %v", i, err)
		}

		d.Remove("", true)
	}

}

func TestMkDirNonExistParent(t *testing.T) {
	tests := []struct {
		diskName string
		fileDir  string
	}{
		// test multiple levels dir
		{"disk0", "dir/sub/way"},
	}

	for i, tt := range tests {
		d := newTestDisk(tt.diskName, "mkdir-non-exist-parent", true)

		err := d.Mkdir(tt.fileDir, false)
		if !os.IsNotExist(err) {
			t.Errorf("%d: expect file not exist, got error = %v", i, err)
		}

		d.Remove("", true)
	}
}

func TestReadEOF(t *testing.T) {
	tests := []struct {
		diskName string
		fileName string
		writeLen int64
		readLen  int64
	}{
		{"disk0", "file", 10, 11},
		{"disk0", "file", 4096 * 2, 4096 * 2 + 1},
	}

	for i, tt := range tests {
		d := newTestDisk(tt.diskName, "mkdir-non-exist-parent", true)
		p := make([]byte, tt.writeLen)
		for i := int64(0); i < tt.writeLen; i++ {
			p[i] = 'X'
		}
		d.WriteAt(tt.fileName, p, 0)
		buf := make([]byte, tt.readLen)
		read, err := d.ReadAt(tt.fileName, buf, 0)
		if int64(read) != tt.writeLen {
			t.Errorf("%d: read %d bytes != %d bytes written", i, read, tt.readLen)
		}
		if err != io.EOF {
			t.Errorf("%d: expect EOF, %v returned", i, err)
		}
		d.Remove("", true)
	}
}
