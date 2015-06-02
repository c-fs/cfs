package main

import (
	"fmt"

	"github.com/c-fs/cfs/client"
	pb "github.com/c-fs/cfs/proto"
	"github.com/qiniu/log"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
)

var (
	mkdirName string
	mkdirAll  bool
)

var mkdirCmd = &cobra.Command{
	Use:   "mkdir",
	Short: "make a dir for a cfs node",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		c := setUpClient()
		defer c.Close()
		handleMkdir(context.TODO(), c)
	},
}

func init() {
	mkdirCmd.PersistentFlags().StringVarP(&mkdirName, "name", "n", "", "dir name")
	mkdirCmd.PersistentFlags().BoolVarP(&mkdirAll, "all", "a", false, "create any necessary parents")
}

func handleMkdir(ctx context.Context, c *client.Client) error {
	err := c.Mkdir(ctx, mkdirName, mkdirAll)
	if err != nil {
		log.Fatalf("Mkdir err (%v)", err)
	}

	fmt.Printf("mkdir %s succeeded\n", mkdirName)
	return nil
}
