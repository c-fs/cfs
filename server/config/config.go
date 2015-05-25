package config

type Server struct {
	Port  string
	Bind  string
	Disks []Disk
}

type Disk struct {
	Name string
	Root string
}
