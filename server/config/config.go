package config

type Server struct {
	Disks []Disk
}

type Disk struct {
	Name string
	Root string
}
