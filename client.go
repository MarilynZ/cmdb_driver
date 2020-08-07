package driver

import (
	"context"

	v1 "github.com/zhihu/cmdb/pkg/api/v1"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

type Driver struct {
	cc *grpc.ClientConn
	v1.ObjectsClient
	handlers []controllerWrapper
}

func NewDriver(addr string) (*Driver, error) {
	clientConn, err := grpc.DialContext(context.Background(), addr, grpc.WithBlock(), grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	cli := v1.NewObjectsClient(clientConn)
	d := &Driver{cc: clientConn, ObjectsClient: cli}
	return d, nil
}

func (d *Driver) RegisterHandler(handler Handler) {
	d.handlers = append(d.handlers, controllerWrapper{
		ObjectsClient: d.ObjectsClient,
		handler:       handler,
	})
}

func (d *Driver) Run(ctx context.Context) (err error) {
	group, ctx := errgroup.WithContext(ctx)

	for _, handler := range d.handlers {
		var h = handler
		group.Go(func() error {
			return h.Run(ctx)
		})
	}
	return group.Wait()
}
