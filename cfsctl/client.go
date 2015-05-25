package main

import (
	"flag"
	"log"

	pb "github.com/c-fs/cfs/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const (
	address = "localhost:15524"
)

func main() {
	// use subcommand instead of action flag
	action := flag.String("action", "", "")

	name := flag.String("name", "", "")
	offset := flag.Int64("offset", 0, "")
	length := flag.Int64("length", 0, "")
	data := flag.String("data", "", "")
	flag.Parse()

	// Set up a connection to the server.
	conn, err := grpc.Dial(address)
	if err != nil {
		log.Fatalf("Cannot connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewCfsClient(conn)

	switch *action {
	case "read":
		reply, err := c.Read(context.TODO(), &pb.ReadRequest{Name: *name, Offset: *offset, Length: *length})
		if err != nil {
			log.Fatalf("Read err (%v)", err)
		}
		log.Println(string(reply.Data))
	case "write":
		reply, err := c.Write(context.TODO(), &pb.WriteRequest{Name: *name, Offset: *offset, Data: []byte(*data)})
		if err != nil {
			log.Fatalf("Write err (%v)", err)
		}
		log.Printf("%d bytes written to %s at offset %d", reply.BytesWritten, *name, *offset)
	default:
		log.Fatalf("Bad action %s", *action)
	}
}
