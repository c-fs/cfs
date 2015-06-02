package main

import (
	"log"

	"github.com/c-fs/cfs/client"
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
	info, err := c.Stats(ctx)
	if err != nil {
		log.Fatalf("ContainerInfo err (%v)", err)
	}
	log.Printf("Container Info: %+v", info)
	return nil
}
