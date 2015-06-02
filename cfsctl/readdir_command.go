package main

import (
	"fmt"

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
		conn := setUpGrpcClient()
		defer conn.Close()
		c := pb.NewCfsClient(conn)

		handleReadDir(context.TODO(), c)
	},
}

func init() {
	readDirCmd.PersistentFlags().StringVarP(&readDirName, "name", "n", "", "readDir name")
}

func handleReadDir(ctx context.Context, c pb.CfsClient) error {
	reply, err := c.ReadDir(ctx, &pb.ReadDirRequest{Name: readDirName})
	if err != nil || reply.Error != nil {
		log.Fatalf("ReadDir err (%v)", err)
	}

	for _, stats := range reply.FileInfos {
		fmt.Printf("%s: %d %t\n", stats.Name, stats.TotalSize, stats.IsDir == true)
	}

	return nil
}
