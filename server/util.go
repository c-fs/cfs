package main

import (
	"fmt"
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
