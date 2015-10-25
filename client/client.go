package client

import (
	"errors"

	pb "github.com/c-fs/cfs/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type Client struct {
	header         *pb.RequestHeader
	grpcConn       *grpc.ClientConn
	metadataClient pb.MetadataClient
	fileClient     pb.CfsClient
	statsClient    pb.StatsClient
}

func New(clientID int64, address string) (*Client, error) {
	header := &pb.RequestHeader{ClientID: clientID}
	conn, err := grpc.Dial(address)
	if err != nil {
		return nil, err
	}
	mc := pb.NewMetadataClient(conn)
	fc := pb.NewCfsClient(conn)
	sc := pb.NewStatsClient(conn)

	return &Client{header: header, grpcConn: conn, metadataClient: mc, fileClient: fc, statsClient: sc}, nil
}

func (c *Client) Disks(ctx context.Context) ([]*pb.Disk, error) {
	reply, err := c.metadataClient.Disks(ctx, &pb.DisksRequest{})
	if err != nil {
		return nil, err
	}
	return reply.Disks, nil
}

func (c *Client) Stat(ctx context.Context, name string) (*pb.FileInfo, error) {
	reply, err := c.fileClient.Stat(ctx, &pb.StatRequest{
		Header: c.header,
		Name:   name,
	})
	if err != nil {
		return nil, err
	}
	return reply.FileInfo, nil
}

func (c *Client) Write(ctx context.Context, name string, offset int64, data []byte, isAppend bool) (int64, error) {
	reply, err := c.fileClient.Write(
		ctx,
		&pb.WriteRequest{Header: c.header, Name: name, Offset: offset, Data: data, Append: isAppend},
	)

	if err != nil {
		return 0, err
	}
	return reply.BytesWritten, parseErr(reply.Error)
}

func (c *Client) Read(ctx context.Context, name string, offset, length int64, checksum uint32,
) (int64, []byte, uint32, error) {
	reply, err := c.fileClient.Read(
		ctx,
		&pb.ReadRequest{
			Header: c.header, Name: name, Offset: offset, Length: length, ExpChecksum: checksum,
		},
	)

	if err != nil {
		return 0, nil, 0, err
	}
	return reply.BytesRead, reply.Data, reply.Checksum, parseErr(reply.Error)
}

func (c *Client) Rename(ctx context.Context, oldName, newName string) error {
	reply, err := c.fileClient.Rename(
		ctx,
		&pb.RenameRequest{Header: c.header, Oldname: oldName, Newname: newName},
	)

	if err != nil {
		return err
	}
	return parseErr(reply.Error)
}

func (c *Client) Remove(ctx context.Context, name string, all bool) error {
	reply, err := c.fileClient.Remove(ctx, &pb.RemoveRequest{Header: c.header, Name: name, All: all})

	if err != nil {
		return err
	}
	return parseErr(reply.Error)
}

func (c *Client) ReadDir(ctx context.Context, name string) ([]*pb.FileInfo, error) {
	reply, err := c.fileClient.ReadDir(ctx, &pb.ReadDirRequest{Header: c.header, Name: name})

	if err != nil {
		return nil, err
	}
	return reply.FileInfos, parseErr(reply.Error)
}

func (c *Client) Mkdir(ctx context.Context, name string, all bool) error {
	reply, err := c.fileClient.Mkdir(ctx, &pb.MkdirRequest{Header: c.header, Name: name, All: all})

	if err != nil {
		return err
	}
	return parseErr(reply.Error)
}

func (c *Client) ContainerInfo(ctx context.Context) (string, error) {
	reply, err := c.statsClient.ContainerInfo(ctx, &pb.ContainerInfoRequest{})

	if err != nil {
		return "", err
	}
	if reply.Error != "" {
		return reply.Info, errors.New(reply.Error)
	}
	return reply.Info, nil
}

func (c *Client) Metrics(ctx context.Context) ([]*pb.Metric, error) {
	reply, err := c.statsClient.Metrics(ctx, &pb.MetricsRequest{})

	if err != nil {
		return nil, err
	}
	return reply.Counters, nil
}

func (c *Client) Close() {
	c.grpcConn.Close()
}

func parseErr(pbErr *pb.Error) error {
	if pbErr == nil {
		return nil
	}
	return errors.New(pbErr.String())
}
