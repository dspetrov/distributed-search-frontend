package clusterManagement

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/go-zookeeper/zk"
)

const (
	WORKERS_REGISTRY_ZNODE      = "/workers_service_registry"
	COORDINATORS_REGISTRY_ZNODE = "/coordinators_service_registry"
)

type ServiceRegistry struct {
	conn                  *zk.Conn
	currentZnode          string
	allServiceAddresses   []string
	serviceRegistryZnode  string
	random                *rand.Rand
	updateAddressesMutex  sync.Mutex
	getRandomAddressMutex sync.Mutex
	getAddressesMutex     sync.Mutex
}

func NewServiceRegistry(conn *zk.Conn, serviceRegistryZnode string) *ServiceRegistry {
	source := rand.NewSource(time.Now().UnixNano())

	se := ServiceRegistry{
		conn:                 conn,
		serviceRegistryZnode: serviceRegistryZnode,
		random:               rand.New(source),
	}

	se.createServiceRegistryNode()

	return &se
}

func (se *ServiceRegistry) RegisterToCluster(metadata string) {
	if se.currentZnode != "" {
		fmt.Println("Already registered to service registry")
		return
	}

	znodePath, err := se.conn.Create(se.serviceRegistryZnode+"/n_", []byte(metadata), zk.FlagEphemeral|zk.FlagSequence, zk.WorldACL(zk.PermAll))
	if err != nil {
		panic(err)
	}

	se.currentZnode = znodePath
	fmt.Println("Registered to service registry")
}

func (se *ServiceRegistry) RegisterForUpdates() {
	se.updateAddresses()
}

func (se *ServiceRegistry) UnregisterFromCluster() {
	if se.currentZnode != "" {
		exists, _, err := se.conn.Exists(se.currentZnode)
		if err != nil {
			panic(err)
		}

		if exists {
			se.conn.Delete(se.currentZnode, -1)
		}
	}
}

func (se *ServiceRegistry) createServiceRegistryNode() {
	registryExists, _, _ := se.conn.Exists(se.serviceRegistryZnode)
	if !registryExists {
		_, err := se.conn.Create(se.serviceRegistryZnode, []byte{}, 0, zk.WorldACL(zk.PermAll))
		if err != nil && err.Error() != "zk: node already exists" {
			panic(err)
		}
	}
}

func (se *ServiceRegistry) GetAllServiceAddresses() []string {
	se.getAddressesMutex.Lock()
	defer se.getAddressesMutex.Unlock()

	if len(se.allServiceAddresses) == 0 {
		se.updateAddresses()
	}

	return se.allServiceAddresses
}

func (se *ServiceRegistry) GetRandomServiceAddress() string {
	se.getRandomAddressMutex.Lock()
	defer se.getRandomAddressMutex.Unlock()

	if len(se.allServiceAddresses) == 0 {
		se.updateAddresses()
	}

	if len(se.allServiceAddresses) > 0 {
		randomIndex := se.random.Intn(len(se.allServiceAddresses))
		return se.allServiceAddresses[randomIndex]
	} else {
		return ""
	}
}

func (se *ServiceRegistry) updateAddresses() {
	se.updateAddressesMutex.Lock()
	defer se.updateAddressesMutex.Unlock()

	workers, _, ch, err := se.conn.ChildrenW(se.serviceRegistryZnode)
	if err != nil {
		panic(err)
	}

	go se.processChildrenEvent(ch)

	addresses := []string{}

	for _, worker := range workers {
		serviceFullpath := se.serviceRegistryZnode + "/" + worker
		exists, _, err := se.conn.Exists(serviceFullpath)
		if err != nil {
			panic(err)
		}
		if !exists {
			continue
		}

		addressBytes, _, err := se.conn.Get(serviceFullpath)
		if err != nil {
			panic(err)
		}

		addresses = append(addresses, string(addressBytes))
	}

	se.allServiceAddresses = addresses
	fmt.Println("The cluster addresses are", se.allServiceAddresses)
}

func (se *ServiceRegistry) processChildrenEvent(ch <-chan zk.Event) {
	event := <-ch
	if event.Type == zk.EventNodeChildrenChanged {
		se.updateAddresses()
	}
}
