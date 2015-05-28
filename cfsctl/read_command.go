package main

import (
	"log"

	pb "github.com/c-fs/cfs/proto"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
)

var (
	readOffset int64
	readName   string
	readLen    int64
)

var readCmd = &cobra.Command{
	Use:   "read",
	Short: "read data from a cfs node",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		rawHandle(context.TODO(), handleRead)
	},
}

func init() {
	readCmd.PersistentFlags().Int64VarP(&readOffset, "offset", "o", 0, "read offset")
	readCmd.PersistentFlags().Int64VarP(&readLen, "length", "l", 0, "read length")
	readCmd.PersistentFlags().StringVarP(&readName, "name", "n", "", "read name")
}

func handleRead(ctx context.Context, c pb.CfsClient) error {
	reply, err := c.Read(
		ctx,
		&pb.ReadRequest{Name: readName, Offset: readOffset, Length: readLen},
	)
	if err != nil {
		log.Fatalf("Read err (%v)", err)
	}
	log.Println(string(reply.Data))

	return nil
}
