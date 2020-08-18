package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"sync"
	"time"

	driver "github.com/cocotyty/cmdb_driver"
	"github.com/cocotyty/cmdb_driver/cache"
	"github.com/cocotyty/cmdb_driver/object"
	v1 "github.com/zhihu/cmdb/pkg/api/v1"
	"github.com/zhihu/cmdb/pkg/signals"
	"google.golang.org/genproto/protobuf/field_mask"
)

const (
	TableOS      = "example_os"
	TableMachine = "example_machine"
)

var addr = flag.String("server", "", "cmdb's address")

func main() {
	flag.Parse()

	d, err := driver.NewDriver(*addr)
	if err != nil {
		log.Fatal(err)
	}
	osCache := &OSCache{
		cache:  map[string]*OS{},
		locker: sync.Mutex{},
	}
	cacheHandler, err := cache.New(TableOS, "", osCache.cache, &osCache.locker)
	if err != nil {
		log.Fatal(err)
	}

	d.RegisterHandler(cacheHandler)
	d.RegisterHandler(&MachineDriver{osCache: osCache})
	ctx := signals.SignalHandler(context.Background())

	err = d.Run(ctx)
	if err != context.Canceled {
		log.Fatal(err)
	}
}

type OSCache struct {
	cache  map[string]*OS
	locker sync.Mutex
}

func (o *OSCache) Get(name string) *OS {
	o.locker.Lock()
	os := o.cache[name]
	o.locker.Unlock()
	return os
}

type OS struct {
	object.Base
	OSVersion string `json:"os_version"`
	Image     string `json:"image"`
}

type Machine struct {
	object.Base
	OS string `json:"os"`
}

const (
	MachineStatusNotReady = "NOT_READY"
	MachineStateBuy       = "BUY"
	MachineStateInited    = "INITED"
	MachineStatusSTABLE   = "STABLE"
	MachineStateNormal    = "NORMAL"
)

type MachineDriver struct {
	client   v1.ObjectsClient
	osCache  *OSCache
	InitJobs []*Machine
	locker   sync.Mutex
}

func (m *MachineDriver) Init(client v1.ObjectsClient) {
	m.client = client
	go m.runLoop()
}

func (m *MachineDriver) Filter() (typ, query string) {
	return TableMachine, ""
}

func (m *MachineDriver) OnUpdate(obj *v1.Object) {
	log.Println(obj)
	// check machine exists?
	var machine = &Machine{}
	err := object.Unmarshal(obj, machine)
	if err != nil {
		log.Println("on update: ", err)
	}
	// if getMachine(machine.Name).Equals(machine){
	// return
	//}
	//
	switch machine.Status {
	case MachineStatusNotReady:
		switch machine.State {
		case MachineStateBuy:
			m.InitMachine(machine)
		case MachineStateInited:
			fmt.Println("do something about this inited machine:", machine.Name)
		}
	default:
		//
	}
}

func (m *MachineDriver) runLoop() {
	timer := time.NewTimer(time.Second * 2)
	for range timer.C {
		m.locker.Lock()
		copied := m.InitJobs
		m.InitJobs = nil
		m.locker.Unlock()
		for _, machine := range copied {
			m.InitMachine(machine)
		}
	}
}

func (m *MachineDriver) InitMachine(machine *Machine) {
	os := m.osCache.Get(machine.OS)
	if os == nil {
		m.locker.Lock()
		m.InitJobs = append(m.InitJobs, machine)
		m.locker.Unlock()
		return
	}
	image := os.Image
	log.Printf("init target machine %s use image: %s", machine.Name, image)
	updated, err := m.client.Update(context.Background(), &v1.ObjectUpdateRequest{
		Object: &v1.Object{
			Type:   machine.Type,
			Name:   machine.Name,
			Status: MachineStatusNotReady,
			State:  MachineStateInited,
		},
		UpdateMask: &field_mask.FieldMask{
			Paths: []string{"state"},
		},
		MatchVersion: false,
	})
	log.Println(updated, err)
}

func (m *MachineDriver) OnDelete(obj *v1.Object) {

}
