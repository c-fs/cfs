package main

import (
	"fmt"
	"io"
	"path"
	"strings"
)

func splitDiskAndFile(name string) (string, string, error) {
	cname := path.Clean(name)
	// min is "a/b"
	// a is disk name
	// b is file name
	if len(cname) < 3 {
		return "", "", fmt.Errorf("bad name: %s", name)
	}

	if cname[0] == '/' {
		cname = cname[1:]
	}
	names := strings.SplitN(cname, "/", 2)
	if len(names) != 2 {
		return "", "", fmt.Errorf("bad name: %s", name)
	}
	return names[0], names[1], nil
}

func NewOffReader(r func([]byte, int64) (int, error)) io.Reader {
	return &offReader{r, 0}
}

func NewOffWriter(w func([]byte, int64) (int, error)) io.Writer {
	return &offWriter{w, 0}
}

type offReader struct {
	reader func([]byte, int64) (int, error)
	off    int64
}

func (or *offReader) Read(b []byte) (int, error) {
	n, err := or.reader(b, or.off)
	or.off += int64(n)
	return n, err
}

type offWriter struct {
	writer func([]byte, int64) (int, error)
	off    int64
}

func (ow *offWriter) Write(b []byte) (int, error) {
	n, err := ow.writer(b, ow.off)
	ow.off += int64(n)
	return n, err
}
