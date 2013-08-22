package cluster

import (
// "encoding/gob"
// "flag"
// "log"
// "os"
// "time"
)

// var nodeConfig struct {
// 	MyNode     Node
// 	ListenAddr string
// }

type NotifyNode struct {
	Node *Node
}

/*type Node struct {
	Name      string
	Env       string
	Version   int
	Address   string
	Host      string
	Pid       int
	BootNanos int64
}

func init() {
	hn, err := os.Hostname()
	if err != nil {
		log.Fatal(err)
	}
	gob.Register(Node{})
	gob.Register(DisGommand{})

	MyHost = hn
	MyNode := &nodeConfig.MyNode
	MyNode.BootNanos = time.Now().UnixNano()
	MyNode.Host = MyHost
	MyNode.Version = DISGO_VERSION
	flag.StringVar(&joinTo, "j", NONE, "Join a specific node")
	flag.StringVar(&bindAddr, "b", "", "Address to bind connections")
	flag.StringVar(&MyNode.Env, "env", "dev", "Operating environment")
	flag.StringVar(&MyNode.Name, "n", DEFAULT_NODE, "Sets the node name")
	flag.StringVar(&nodeConfig.ListenAddr, "l", ":8765", "TCP port to listen on")

	register = make(chan *NodeConnection)
	unregister = make(chan string)
	go registryLoop()

}

func startNode() {
	initListener()
	// dbstuff.Dodbstuff()
	if joinTo != NONE {
		err := attemptJoin(joinTo)
		if err != nil {
			log.Fatal(err)
		}
	}
	waitLock.Add(1)
	log.Println("Waiting for exit")
	waitLock.Wait()
}
*/
