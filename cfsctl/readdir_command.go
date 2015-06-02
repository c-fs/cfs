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
	readDirName string
)

var readDirCmd = &cobra.Command{
	Use:   "readDir",
	Short: "readDir from a cfs node",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		c := setUpClient()
		defer c.Close()

		handleReadDir(context.TODO(), c)
	},
}

func init() {
	readDirCmd.PersistentFlags().StringVarP(&readDirName, "name", "n", "", "readDir name")
}

func handleReadDir(ctx context.Context, c *client.Client) error {
	fInfos, err := c.ReadDir(ctx, readDirName)
	if err != nil {
		log.Fatalf("ReadDir err (%v)", err)
	}

	for _, stats := range fInfos {
		fmt.Printf("%s: %d %t\n", stats.Name, stats.TotalSize, stats.IsDir == true)
	}

	return nil
}
