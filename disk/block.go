package disk

import (
	"encoding/binary"
	"errors"
	"hash/crc32"
	"io"
	"os"
)

var (
	crc32cTable       = crc32.MakeTable(crc32.Castagnoli)
	crc32Len          = int64(4)
	// ErrPayloadSizeTooLarge indicates the input payload size is too big
	ErrPayloadSizeTooLarge = errors.New("disk: bad payload size")
	// ErrBadCRC indicates there is not CRC can be found in the block
	ErrBadCRC = errors.New("disk: not a valid CRC")
)

// readBlock reads a full or partial block into p from rs. len(p) must be smaller than (bs - crc32Len).
// The checksum of the block is calculated and verified.
// ErrBadCRC is returned if the checksum is invalid.
func readBlock(rs io.ReadSeeker, p []byte, index, bs int64) (int64, error) {
	if int64(len(p)) > bs-crc32Len {
		return 0, ErrPayloadSizeTooLarge
	}
	b := make([]byte, crc32Len)
	rs.Seek(index*bs, os.SEEK_SET)
	n, err := rs.Read(b)
	// Cannot read full crc
	if n > 0 && n < 4 {
		return 0, ErrBadCRC
	}
	if err != nil {
		return 0, err
	}

	n, err = rs.Read(p)
	if err != nil && err != io.EOF {
		return 0, err
	}

	crc := binary.BigEndian.Uint32(b)
	// Invalid crc
	if crc != crc32.Checksum(p[:n], crc32cTable) {
		return 0, ErrBadCRC
	}
	return int64(n) - crc32Len, nil
}

// writeBlock writes a full or partial block into ws. len(p) must be smaller than (bs - crc32Len).
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
