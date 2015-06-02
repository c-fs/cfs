package main

import (
	"github.com/c-fs/cfs/client"
	"github.com/qiniu/log"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
)

var (
	readOffset      int64
	readName        string
	readLen         int64
	readExpChecksum uint32
)

var readCmd = &cobra.Command{
	Use:   "read",
	Short: "read data from a cfs node",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		c := setUpClient()
		defer c.Close()

		handleRead(context.TODO(), c)
	},
}

func init() {
	readCmd.PersistentFlags().Int64VarP(&readOffset, "offset", "o", 0, "read offset")
	readCmd.PersistentFlags().Int64VarP(&readLen, "length", "l", 0, "read length")
	readCmd.PersistentFlags().StringVarP(&readName, "name", "n", "", "read name")
	readCmd.PersistentFlags().Uint32VarP(&readExpChecksum, "checksum", "c", 0, "expect checksum")
}

func handleRead(ctx context.Context, c *client.Client) error {
	_, data, _, err := c.Read(ctx, readName, readOffset, readLen, readExpChecksum)
	if err != nil {
		log.Fatalf("Read err (%v)", err)
	}
	log.Println(string(data))

	return nil
}
