package main

import (
	"github.com/c-fs/cfs/client"
	"github.com/qiniu/log"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
)

var (
	removeName string
	removeAll  bool
)

var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "remove file from a cfs node",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		c := setUpClient()
		defer c.Close()

		handleRemove(context.TODO(), c)
	},
}

func init() {
	removeCmd.PersistentFlags().StringVarP(&removeName, "name", "n", "", "remove file name")
	removeCmd.PersistentFlags().BoolVarP(&removeAll, "all", "a", false, "remove all files")
}

func handleRemove(ctx context.Context, c *client.Client) error {
	err := c.Remove(ctx, removeName, removeAll)
	if err != nil {
		log.Fatalf("Read err (%v)", err)
	}
	log.Info("remove succeeded")

	return nil
}
