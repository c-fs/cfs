package main

import (
	pb "github.com/c-fs/cfs/proto"
	"github.com/qiniu/log"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
)

var (
	writeName   string
	writeData   string
	writeOffset int64
)

var writeCmd = &cobra.Command{
	Use:   "write",
	Short: "write data to a cfs node",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		conn := setUpGrpcClient()
		defer conn.Close()
		c := pb.NewCfsClient(conn)

		handleWrite(context.TODO(), c)
	},
}

func init() {
	writeCmd.PersistentFlags().StringVarP(&writeName, "name", "n", "", "write name")
	writeCmd.PersistentFlags().Int64VarP(&writeOffset, "offset", "o", 0, "write offset")
	writeCmd.PersistentFlags().StringVarP(&writeData, "data", "d", "", "write data")
}

func handleWrite(ctx context.Context, c pb.CfsClient) error {
	reply, err := c.Write(
		ctx,
		&pb.WriteRequest{Name: writeName, Offset: writeOffset, Data: []byte(writeData)},
	)
	if err != nil {
		log.Fatalf("Write err (%v)", err)
	}
	log.Printf("%d bytes written to %s at offset %d",
		reply.BytesWritten, writeName, writeOffset)

	return nil
}
