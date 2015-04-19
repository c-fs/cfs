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
	// ErrBadPayloadSize indicates the input payload size is invalid
	ErrBadPayloadSize = errors.New("disk: bad payload size")
	// ErrInvalidCRC indicates there is not CRC can be found in the block
	ErrInvalidCRC = errors.New("disk: not a valid CRC")
)

// ReadBlock reads a full block into p from rs. len(p) must be smaller than (bs - crc32Len).
// The checksum of the block is calculated and verified.
// ErrInvalidCRC is returned if the checksum is invalid.
func ReadBlock(rs io.ReadSeeker, p []byte, index, bs int64 ) error {
	if int64(len(p)) != bs-crc32Len {
		return ErrBadPayloadSize
	}
	buf := make([]byte, bs)
	rs.Seek(index*bs, os.SEEK_SET)
	n, err := rs.Read(buf)
	if err != nil && err != io.EOF {
		return err
	}
	// Cannot read crc
	if n < 4 {
		return ErrInvalidCRC
	}
	payload := buf[crc32Len:]
	crc := binary.BigEndian.Uint32(buf[:crc32Len])
	// Invalid crc
	if crc != crc32.Checksum(payload, crc32cTable) {
		return ErrInvalidCRC
	}
	copy(p, payload)
	return nil
}

// writeBlock writes a full or partial block into ws. len(p) must be smaller than (bs - crc32Len).
// Each block contains a crc32Len bytes crc32c of the payload and a (bs-crc32Len)
// bytes payload. Any error encountered is returned.
// It is caller's responsibility to ensure that only the last block in the file has
// len(p) < bs - crc32Len
func writeBlock(ws io.WriteSeeker, index, bs int64, p []byte) error {
	if int64(len(p)) > bs-crc32Len {
		return ErrBadPayloadSize
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
