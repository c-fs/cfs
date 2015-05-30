package main

import (
	"fmt"
	"log"

	pb "github.com/c-fs/cfs/proto"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
)

var (
	mkdirName string
	mkdirAll  bool
)

var mkdirCmd = &cobra.Command{
	Use:   "mkdir",
	Short: "make a dir for a cfs node",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		conn := setUpGrpcClient()
		defer conn.Close()
		c := pb.NewCfsClient(conn)

		handleMkdir(context.TODO(), c)
	},
}

func init() {
	mkdirCmd.PersistentFlags().StringVarP(&mkdirName, "name", "n", "", "dir name")
	mkdirCmd.PersistentFlags().BoolVarP(&mkdirAll, "all", "a", false, "create any necessary parents")
}

func handleMkdir(ctx context.Context, c pb.CfsClient) error {
	req, err := c.Mkdir(ctx, &pb.MkdirRequest{Name: mkdirName, All: mkdirAll})
	if err != nil || req.Error != nil {
		log.Fatalf("Mkdir err (%v)", err)
	}

	fmt.Printf("mkdir %s succeeded\n", mkdirName)
	return nil
}
