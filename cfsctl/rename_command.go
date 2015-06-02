package main

import (
	pb "github.com/c-fs/cfs/proto"
	"github.com/qiniu/log"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
)

var (
	renameOld string
	renameNew string
)

var renameCmd = &cobra.Command{
	Use:   "rename",
	Short: "rename a file on a cfs node",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		conn := setUpGrpcClient()
		defer conn.Close()
		c := pb.NewCfsClient(conn)

		handleRename(context.TODO(), c)
	},
}

func init() {
	renameCmd.PersistentFlags().StringVarP(&renameOld, "oldname", "", "", "old name")
	renameCmd.PersistentFlags().StringVarP(&renameNew, "newname", "", "", "new name")
}

func handleRename(ctx context.Context, c pb.CfsClient) error {
	_, err := c.Rename(
		ctx,
		&pb.RenameRequest{Oldname: renameOld, Newname: renameNew},
	)
	if err != nil {
		log.Fatalf("Rename err (%v)", err)
	}
	log.Printf("rename %s into %s", renameOld, renameNew)

	return nil
}
