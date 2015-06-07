package main

import (
	"github.com/c-fs/cfs/client"
	"github.com/qiniu/log"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "display cfs stats",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		c := setUpClient()
		defer c.Close()

		handleStats(context.TODO(), c)
	},
}

func handleStats(ctx context.Context, c *client.Client) error {
	info, err := c.ContainerInfo(ctx)
	if err != nil {
		log.Printf("ContainerInfo err (%v)", err)
	} else {
		log.Printf("Container Info: %+v", info)
	}
	ms, err := c.Metrics(ctx)
	if err != nil {
		log.Printf("Metrics err (%v)", err)
	} else {
		log.Printf("Metrics: %+v", ms)
	}
	return nil
}
