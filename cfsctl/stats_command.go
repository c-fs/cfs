package main

import (
	pb "github.com/c-fs/cfs/proto"
	"github.com/qiniu/log"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "display cfs stats",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		conn := setUpGrpcClient()
		defer conn.Close()
		c := pb.NewStatsClient(conn)

		handleStats(context.TODO(), c)
	},
}

func handleStats(ctx context.Context, c pb.StatsClient) error {
	reply, err := c.ContainerInfo(context.TODO(), &pb.ContainerInfoRequest{})
	if err != nil || reply.Error != "" {
		log.Fatalf("ContainerInfo err (%v)", err)
	}
	log.Printf("Container Info: %+v", reply.Info)
	return nil
}
