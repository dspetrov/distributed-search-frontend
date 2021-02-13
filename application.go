package main

import (
	"dspetrov/distributed-search-frontend/clusterManagement"
	"dspetrov/distributed-search-frontend/networking"
	"dspetrov/distributed-search-frontend/search"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/go-zookeeper/zk"
)

const (
	ZOOKEEPER_ADDRESS = "localhost:2181"
	SESSION_TIMEOUT   = 3000
	DEFAULT_PORT      = 9000
)

var conn *zk.Conn

func main() {
	var currentServerPort int
	if len(os.Args) == 2 {
		currentServerPort, _ = strconv.Atoi(os.Args[1])
	} else {
		currentServerPort = DEFAULT_PORT
	}

	conn := connectToZookeeper()

	coordinatorsServiceRegistry := clusterManagement.NewServiceRegistry(conn, clusterManagement.COORDINATORS_REGISTRY_ZNODE)

	searchHandler := search.NewUserSearchHandler(coordinatorsServiceRegistry)
	webServer := networking.NewWebServer(currentServerPort, searchHandler)
	go webServer.StartServer()

	fmt.Println("Server is listening on port", currentServerPort)

	run()
	close()

	fmt.Println("Disconnected from Zookeeper, exiting application")
}

func connectToZookeeper() *zk.Conn {
	c, ch, err := zk.Connect([]string{ZOOKEEPER_ADDRESS}, SESSION_TIMEOUT*time.Millisecond)
	if err != nil {
		panic(err)
	}

	conn = c
	go processConnectEvent(ch)

	return c
}

func run() {
	<-make(chan int)
	// _, _, ch, err := conn.ChildrenW("/")
	// if err != nil {
	// 	panic(err)
	// }

	// <-ch
}

func close() {
	conn.Close()
}

func processConnectEvent(ch <-chan zk.Event) {
	for event := range ch {
		if event.Type == zk.EventSession {
			switch event.State {
			case zk.StateConnected:
				fmt.Println("Successfully connected to ZooKeeper")
			case zk.StateDisconnected:
				fmt.Println("Disconnected from ZooKeeper event")
				// le.conn ??? Notify()
			}
		}
	}
}
