// +build !linux

package main

import "fmt"

func detectDockerContainer() (string, error) {
	return "", fmt.Errorf("unsupported to detect docker container")
}
