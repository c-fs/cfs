// +build linux

package stats

import (
	"time"

	"github.com/google/cadvisor/info/v1"
	"github.com/google/cadvisor/manager"
	"github.com/google/cadvisor/storage/memory"
	"github.com/google/cadvisor/utils/sysfs"
	"github.com/qiniu/log"
)

const (
	statsToCacheNum = 60
	storageDuration = 2 * time.Minute
)

var cmgr manager.Manager

func initContainerManager() {
	sysFs, err := sysfs.NewRealSysFs()
	if err != nil {
		log.Infof("stats: failed to create a system interface (%v)", err)
		return
	}
	// TODO: support influxdb or other backend storage
	cmgr, err = manager.New(memory.New(storageDuration, nil), sysFs)
	if err != nil {
		log.Infof("stats: failed to create a container Manager (%v)", err)
		return
	}
	if err := cmgr.Start(); err != nil {
		log.Infof("stats: failed to start container manager (%v)", err)
		return
	}
}

func containerInfo() (*v1.ContainerInfo, error) {
	infoReq := v1.DefaultContainerInfoRequest()
	return cmgr.GetContainerInfo(containerName, &infoReq)
}
