package object

import (
	"errors"
	"reflect"
	"strconv"
	"time"

	"github.com/cocotyty/forceset"
	"github.com/golang/protobuf/ptypes"
	v1 "github.com/zhihu/cmdb/pkg/api/v1"
)

var (
	ErrTypeNotSupport = errors.New("type not support")
)

type Base struct {
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	State       string    `json:"state"`
	Status      string    `json:"status"`
	Description string    `json:"description"`
	Version     uint64    `json:"version"`
	CreateTime  time.Time `json:"create_time"`
}

func Unmarshal(obj *v1.Object, target interface{}) (err error) {
	ptr := reflect.ValueOf(target)
	if ptr.Kind() != reflect.Ptr {
		return ErrTypeNotSupport
	}
	st := ptr.Elem()
	if st.Kind() != reflect.Struct {
		return ErrTypeNotSupport
	}
	metas := map[string]interface{}{}
	for name, mv := range obj.Metas {
		switch mv.ValueType {
		case v1.ValueType_STRING:
			metas[name] = mv.Value
		case v1.ValueType_BOOLEAN:
			b, _ := strconv.ParseBool(mv.Value)
			metas[name] = b
		case v1.ValueType_DOUBLE:
			f, _ := strconv.ParseFloat(mv.Value, 64)
			metas[name] = f
		case v1.ValueType_INTEGER:
			i, _ := strconv.ParseInt(mv.Value, 10, 64)
			metas[name] = i
		}
	}

	err = forceset.Set(target, metas)
	if err != nil {
		return err
	}
	b := Base{
		Name:        obj.Name,
		Type:        obj.Type,
		State:       obj.State,
		Status:      obj.Status,
		Description: obj.Description,
		Version:     obj.Version,
	}
	if obj.CreateTime != nil {
		b.CreateTime, _ = ptypes.Timestamp(obj.CreateTime)
	}
	err = forceset.Set(target, obj)
	return err
}
