package main

import (
	"github.com/c-fs/cfs/client"
	pb "github.com/c-fs/cfs/proto"
	"github.com/qiniu/log"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
)

var (
	renameOld string
	renameNew string
)

var renameCmd = &cobra.Command{
	Use:   "rename",
	Short: "rename a file on a cfs node",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		c := setUpClient()
		defer c.Close()

		handleRename(context.TODO(), c)
	},
}

func init() {
	renameCmd.PersistentFlags().StringVarP(&renameOld, "oldname", "", "", "old name")
	renameCmd.PersistentFlags().StringVarP(&renameNew, "newname", "", "", "new name")
}

func handleRename(ctx context.Context, c *client.Client) error {
	err := c.Rename(ctx, renameOld, renameNew)
	if err != nil {
		log.Fatalf("Rename err (%v)", err)
	}
	log.Printf("rename %s into %s", renameOld, renameNew)

	return nil
}
