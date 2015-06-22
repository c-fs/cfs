package main

import (
	"github.com/c-fs/cfs/client"
	"github.com/qiniu/log"
	"github.com/spf13/cobra"
)

var (
	address string
)

var cfsctlCmd = &cobra.Command{
	Use:   "cfsctl",
	Short: "cfsctl is cfs comment line client",
	Long: `cfsctl is the main command, used to communicate to your cfs node.

cfs is the lowest building block of cloud infrastructure that is available on all machines.

Complete documentation is available at https://github.com/c-fs/cfs`,
	Run: nil,
}

func init() {
	cfsctlCmd.PersistentFlags().StringVarP(&address, "address", "",
		"localhost:15524", "address of the cfs node server")
	addCommand()
}

func addCommand() {
	cfsctlCmd.AddCommand(readCmd)
	cfsctlCmd.AddCommand(writeCmd)
	cfsctlCmd.AddCommand(renameCmd)
	cfsctlCmd.AddCommand(removeCmd)
	cfsctlCmd.AddCommand(readDirCmd)
	cfsctlCmd.AddCommand(mkdirCmd)
	cfsctlCmd.AddCommand(statsCmd)
}

func setUpClient() *client.Client {
	// Set up a connection to the server.
	c, err := client.New(0x1234, address)
	if err != nil {
		log.Fatalf("Cannot create cfs client: %v", err)
	}
	return c
}
