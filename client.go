package driver

import (
	"context"

	v1 "github.com/zhihu/cmdb/pkg/api/v1"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

type Driver struct {
	cc                  *grpc.ClientConn
	ObjectsClient       v1.ObjectsClient
	RelationsClient     v1.RelationsClient
	ObjectTypesClient   v1.ObjectTypesClient
	RelationTypesClient v1.RelationTypesClient
	handlers            []controllerWrapper
}

func NewDriver(addr string) (*Driver, error) {
	clientConn, err := grpc.DialContext(context.Background(), addr, grpc.WithBlock(), grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	d := &Driver{
		cc:                  clientConn,
		ObjectsClient:       v1.NewObjectsClient(clientConn),
		RelationsClient:     v1.NewRelationsClient(clientConn),
		ObjectTypesClient:   v1.NewObjectTypesClient(clientConn),
		RelationTypesClient: v1.NewRelationTypesClient(clientConn),
	}
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
