package command

import (
	"log"

	pb "github.com/c-fs/cfs/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const (
	address = "localhost:15524"
)

type handlerFunc func(context.Context, pb.CfsClient) error

func rawHandle(ctx context.Context, fn handlerFunc) {
	// Set up a connection to the server.
	conn, err := grpc.Dial(address)
	if err != nil {
		log.Fatalf("Cannot connect: %v", err)
	}
	defer conn.Close()

	c := pb.NewCfsClient(conn)

	fn(ctx, c)
}
