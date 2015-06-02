// +build !linux

package stats

import (
	"fmt"

	"github.com/google/cadvisor/info/v1"
)

func initContainerManager() {}

func containerInfo() (*v1.ContainerInfo, error) {
	return nil, fmt.Errorf("unsupported to collect container info")
}
