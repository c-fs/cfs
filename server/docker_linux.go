package main

import "github.com/docker/libcontainer/cgroups"

func detectDockerContainer() (string, error) {
	return cgroups.GetThisCgroupDir("cpu")
}
