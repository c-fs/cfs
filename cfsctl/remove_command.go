package main

import (
	pb "github.com/c-fs/cfs/proto"
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
		conn := setUpGrpcClient()
		defer conn.Close()
		c := pb.NewCfsClient(conn)

		handleRemove(context.TODO(), c)
	},
}

func init() {
	removeCmd.PersistentFlags().StringVarP(&removeName, "name", "n", "", "remove file name")
	removeCmd.PersistentFlags().BoolVarP(&removeAll, "all", "a", false, "remove all files")
}

func handleRemove(ctx context.Context, c pb.CfsClient) error {
	_, err := c.Remove(
		ctx,
		&pb.RemoveRequest{Name: removeName, All: removeAll},
	)
	if err != nil {
		log.Fatalf("Read err (%v)", err)
	}
	log.Println("deletion succeeded")

	return nil
}
