package cache

import (
	"errors"
	"log"
	"reflect"
	"sync"

	"github.com/cocotyty/cmdb_driver/object"
	v1 "github.com/zhihu/cmdb/pkg/api/v1"
)

type Cache struct {
	m         interface{} // a map to store
	mValue    reflect.Value
	locker    sync.Locker // locker
	typ       string
	query     string
	valueType reflect.Type
}

var UnsupportedType = errors.New("store must be a map<string,type>")

func New(typ string, query string, store interface{}, locker sync.Locker) (c *Cache, err error) {
	mValue := reflect.ValueOf(store)
	mType := mValue.Type()
	if mType.Kind() != reflect.Map {
		return nil, UnsupportedType
	}
	if mType.Key().Kind() != reflect.String {
		return nil, UnsupportedType
	}
	c = &Cache{
		m:         store,
		mValue:    mValue,
		locker:    locker,
		typ:       typ,
		query:     query,
		valueType: mType.Elem(),
	}
	return c, nil
}

func (c *Cache) Init(_ v1.ObjectsClient) {}

func (c *Cache) OnUpdate(obj *v1.Object) {
	value := reflect.New(c.valueType.Elem())
	err := object.Unmarshal(obj, value.Interface())
	if err != nil {
		log.Println("[ERROR] [Cache] OnUpdate", err)
		return
	}
	log.Println(obj)
	c.locker.Lock()
	c.mValue.SetMapIndex(reflect.ValueOf(obj.Name), value)
	c.locker.Unlock()
}

func (c *Cache) OnDelete(obj *v1.Object) {
	value := reflect.New(c.valueType)
	err := object.Unmarshal(obj, value.Elem().Interface())
	if err != nil {
		log.Println("[ERROR] [Cache] OnUpdate", err)
		return
	}
	c.locker.Lock()
	c.mValue.SetMapIndex(reflect.ValueOf(obj.Name), reflect.Value{})
	c.locker.Unlock()
}

func (c *Cache) Filter() (typ, query string) {
	return c.typ, c.query
}
