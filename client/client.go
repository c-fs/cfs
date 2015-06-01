package client

import (
	"errors"

	pb "github.com/c-fs/cfs/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type Client struct {
	grpcConn   *grpc.ClientConn
	grpcClient pb.CfsClient
}

func New(address string) (*Client, error) {
	conn, err := grpc.Dial(address)
	if err != nil {
		return nil, err
	}
	pbClient := pb.NewCfsClient(conn)
	return &Client{grpcConn: conn, grpcClient: pbClient}, nil
}

func (c *Client) Write(ctx context.Context, name string, offset int64, data []byte, isAppend bool) (int64, error) {
	reply, err := c.grpcClient.Write(
		ctx,
		&pb.WriteRequest{Name: name, Offset: offset, Data: data, Append: isAppend},
	)

	if err != nil {
		return 0, err
	}
	return reply.BytesWritten, parseErr(reply.Error)
}

func (c *Client) Read(ctx context.Context, name string, offset, length int64, checkSum uint32,
) (int64, []byte, uint32, error) {
	reply, err := c.grpcClient.Read(
		ctx,
		&pb.ReadRequest{
			Name: name, Offset: offset, Length: length, ExpChecksum: checkSum,
		},
	)

	if err != nil {
		return 0, nil, 0, err
	}
	return reply.BytesRead, reply.Data, reply.Checksum, parseErr(reply.Error)
}

func (c *Client) Rename(ctx context.Context, oldName, newName string) error {
	reply, err := c.grpcClient.Rename(
		ctx,
		&pb.RenameRequest{Oldname: oldName, Newname: newName},
	)

	if err != nil {
		return err
	}
	return parseErr(reply.Error)
}

func (c *Client) Remove(ctx context.Context, name string, all bool) error {
	reply, err := c.grpcClient.Remove(ctx, &pb.RemoveRequest{Name: name, All: all})

	if err != nil {
		return err
	}
	return parseErr(reply.Error)
}

func (c *Client) ReadDir(ctx context.Context, name string) ([]*pb.FileInfo, error) {
	reply, err := c.grpcClient.ReadDir(ctx, &pb.ReadDirRequest{Name: name})

	if err != nil {
		return nil, err
	}
	return reply.FileInfos, parseErr(reply.Error)
}

func (c *Client) Mkdir(ctx context.Context, name string, all bool) error {
	reply, err := c.grpcClient.Mkdir(ctx, &pb.MkdirRequest{Name: name, All: all})

	if err != nil {
		return err
	}
	return parseErr(reply.Error)
}

func (c *Client) Close() {
	c.grpcConn.Close()
}

func parseErr(pbErr *pb.Error) error {
	return errors.New(pbErr.String())
}
