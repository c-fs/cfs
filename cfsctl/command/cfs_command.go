package command

import "github.com/spf13/cobra"

var (
	address string
)

var CfsctlCmd = &cobra.Command{
	Use:   "cfsctl",
	Short: "cfsctl is cfs comment line client",
	Long: `cfsctl is the main command, used to communicate to your cfs node.

cfs is the lowest building block of cloud infrastructure that is available on all machines.

Complete documentation is available at https://github.com/c-fs/cfs`,
	Run: nil,
}

func init() {
	CfsctlCmd.PersistentFlags().StringVarP(&address, "address", "",
		"localhost:15524", "address of the cfs node server")
	addCommand()
}

func addCommand() {
	CfsctlCmd.AddCommand(readCmd)
	CfsctlCmd.AddCommand(writeCmd)
	CfsctlCmd.AddCommand(renameCmd)
	CfsctlCmd.AddCommand(removeCmd)
}
