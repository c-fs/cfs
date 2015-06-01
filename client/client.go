package client

import (
	"errors"
	"log"

	pb "github.com/c-fs/cfs/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type Client struct {
	conn *grpc.ClientConn
}

func New(address string) *Client {
	conn, err := grpc.Dial(address)
	if err != nil {
		log.Fatalf("Cannot connect: %v", err)
	}
	return &Client{conn: conn}
}

func (c *Client) Write(
	ctx context.Context, name string, offset int64, data []byte, isAppend bool,
) (int64, error) {
	pbClient := pb.NewCfsClient(c.conn)
	reply, err := pbClient.Write(
		ctx,
		&pb.WriteRequest{Name: name, Offset: offset, Data: data, Append: isAppend},
	)

	if err != nil {
		return 0, err
	}
	return reply.BytesWritten, pbErr2GoErr(reply.Error)
}

func (c *Client) Read(
	ctx context.Context, name string, offset, length int64, checkSum uint32,
) (int64, []byte, uint32, error) {
	pbClient := pb.NewCfsClient(c.conn)
	reply, err := pbClient.Read(
		ctx,
		&pb.ReadRequest{
			Name: name, Offset: offset, Length: length, ExpChecksum: checkSum,
		},
	)

	if err != nil {
		return 0, nil, 0, err
	}
	return reply.BytesRead, reply.Data, reply.Checksum, pbErr2GoErr(reply.Error)
}

func (c *Client) Rename(ctx context.Context, oldName, newName string) error {
	pbClient := pb.NewCfsClient(c.conn)
	reply, err := pbClient.Rename(
		ctx,
		&pb.RenameRequest{Oldname: oldName, Newname: newName},
	)

	if err != nil {
		return err
	}
	return pbErr2GoErr(reply.Error)
}

func (c *Client) Remove(ctx context.Context, name string, all bool) error {
	pbClient := pb.NewCfsClient(c.conn)
	reply, err := pbClient.Remove(ctx, &pb.RemoveRequest{Name: name, All: all})

	if err != nil {
		return err
	}
	return pbErr2GoErr(reply.Error)
}

func (c *Client) ReadDir(ctx context.Context, name string) ([]*pb.FileInfo, error) {
	pbClient := pb.NewCfsClient(c.conn)
	reply, err := pbClient.ReadDir(ctx, &pb.ReadDirRequest{Name: name})

	if err != nil {
		return nil, err
	}
	return reply.FileInfos, pbErr2GoErr(reply.Error)
}

func (c *Client) Mkdir(ctx context.Context, name string, all bool) error {
	pbClient := pb.NewCfsClient(c.conn)
	reply, err := pbClient.Mkdir(ctx, &pb.MkdirRequest{Name: name, All: all})

	if err != nil {
		return err
	}
	return pbErr2GoErr(reply.Error)
}

func (c *Client) Close() {
	c.conn.Close()
}

func pbErr2GoErr(pbErr *pb.Error) error {
	return errors.New(pbErr.String())
}
