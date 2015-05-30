package main

import (
	"log"

	pb "github.com/c-fs/cfs/proto"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
)

var (
	syncName string
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "sync files to cfs node",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		conn := setUpGrpcClient()
		defer conn.Close()
		c := pb.NewCfsClient(conn)

		handleSync(context.TODO(), c)
	},
}

func init() {
	syncCmd.PersistentFlags().StringVarP(&syncName, "name", "n", "", "sync name")
}

func handleSync(ctx context.Context, c pb.CfsClient) error {
	reply, err := c.Sync(ctx, &pb.SyncRequest{Names: []string{syncName}})
	if err != nil || reply.Errors != nil {
		log.Fatalf("Sync %s err (%v)", syncName, err)
	}

	log.Printf("Sync %s succeeded", syncName)
	return nil
}
