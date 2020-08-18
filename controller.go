package driver

import (
	"context"

	v1 "github.com/zhihu/cmdb/pkg/api/v1"
)

type Handler interface {
	Init(client v1.ObjectsClient)
	Filter() (typ, query string)
	OnUpdate(object *v1.Object)
	OnDelete(object *v1.Object)
}

type controllerWrapper struct {
	v1.ObjectsClient
	handler Handler
}

func (c *controllerWrapper) Run(ctx context.Context) error {
	typ, query := c.handler.Filter()

	watchClient, err := c.Watch(ctx, &v1.ListObjectRequest{
		Type:        typ,
		View:        v1.ObjectView_NORMAL,
		Query:       query,
		ShowDeleted: false,
	})
	if err != nil {
		return err
	}
	c.handler.Init(c.ObjectsClient)
	for {
		evt, err := watchClient.Recv()
		if err != nil {
			return err
		}
		c.Receive(evt)
	}
}

func (c *controllerWrapper) Receive(evt *v1.ObjectEvent) {
	switch evt.GetType() {
	case v1.WatchEventType_INIT, v1.WatchEventType_UPDATE, v1.WatchEventType_CREATE:
		for _, obj := range evt.Objects {
			c.handler.OnUpdate(obj)
		}
	case v1.WatchEventType_DELETE:
		for _, obj := range evt.Objects {
			c.handler.OnDelete(obj)
		}
	}
}
