package main

import (
	"fmt"

	"github.com/c-fs/cfs/client"
	"github.com/qiniu/log"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
)

var disksCmd = &cobra.Command{
	Use:   "disks",
	Short: "get available disk names",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		c := setUpClient()
		defer c.Close()

		handleDisks(context.TODO(), c)
	},
}

func handleDisks(ctx context.Context, c *client.Client) error {
	disks, err := c.Disks(ctx)
	if err != nil {
		log.Fatalf("Disks err (%v)", err)
	}
	for _, d := range disks {
		fmt.Println(d.Name)
	}
	return nil
}
