package main

import (
	"fmt"

	"github.com/c-fs/cfs/client"
	"github.com/qiniu/log"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
)

var (
	src string
	dst string
)

var copyCmd = &cobra.Command{
	Use:   "copy",
	Short: "copy a file",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		c := setUpClient()
		defer c.Close()
		handleCopy(context.TODO(), c)
	},
}

func init() {
	copyCmd.PersistentFlags().StringVarP(&src, "src", "s", "", "source file")
	copyCmd.PersistentFlags().StringVarP(&dst, "dst", "d", "", "target file")
}

func handleCopy(ctx context.Context, c *client.Client) error {
	if err := c.Copy(ctx, src, dst); err != nil {
		log.Fatal("copy failed -", err)
	}

	fmt.Printf("copy %s => %s succeeded\n", src, dst)
	return nil
}
