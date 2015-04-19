package disk

import "io"

func ReadBlock(file io.ReadSeeker, index, bs int64) ([]byte, error) {

}

func WriteBlock(file io.WriteSeeker, index, bs int64, data []byte) error {

}
