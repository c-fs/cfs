package disk

import (
	"encoding/binary"
	"errors"
	"hash/crc32"
	"io"
	"os"
)

var (
	crc32cTable            = crc32.MakeTable(crc32.Castagnoli)
	crc32Len               = int64(4)
	ErrPayloadSizeTooLarge = errors.New("disk: payload size too large")
)

// writeBlock writes a full block into ws. len(p) must be smaller than (bs - crc32Len).
// Each block contains a crc32Len bytes crc32c of the payload and a (bs-crc32Len)
// bytes payload. Any error encountered is returned.
// It is caller's responsibility to ensure that only the last block in the file has
// len(p) < bs - crc32Len
func writeBlock(ws io.WriteSeeker, index, bs int64, p []byte) error {
	if int64(len(p)) > bs-crc32Len {
		return ErrPayloadSizeTooLarge
	}

	// seek to the beginning of the block
	_, err := ws.Seek(index*bs, os.SEEK_SET)
	if err != nil {
		return nil
	}

	// write crc32c
	// TODO: reuse buffer
	b := make([]byte, crc32Len)
	binary.BigEndian.PutUint32(b, crc32.Checksum(p, crc32cTable))
	_, err = ws.Write(b)
	if err != nil {
		return err
	}

	// write payload
	_, err = ws.Write(p)
	if err != nil {
		return err
	}
	return nil
}
