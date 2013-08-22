package main

import (
	"flag"
	"github.com/peak6/disgo/cluster"
	"github.com/peak6/disgo/om"
	log_pkg "log"
	"os"
	// "sync"
)

const (
	DISGO_VERSION = 1
)

const (
	NONE         string = "none"
	DEFAULT_NODE string = "<ENV>"
)

/*var waitLock sync.WaitGroup
var MyHost string
var joinTo string
var bindAddr string
var udpPort int
var tcpPort int
var register chan *NodeConnection
var unregister chan string
*/
var maps map[string]*om.OwnerMap
var log = log_pkg.New(os.Stderr, "[disgo] ", log_pkg.LstdFlags|log_pkg.Lshortfile|log_pkg.Lmicroseconds)

func init() {
	maps = make(map[string]*om.OwnerMap)
}

func main() {
	log_pkg.SetFlags(log_pkg.LstdFlags | log_pkg.Lshortfile | log_pkg.Lmicroseconds)
	flag.Parse()
	// runtime.GOMAXPROCS(runtime.NumCPU())
	// var mapCount = 5
	// tgt := om.New()

	// maps := []*om.OwnerMap{}
	// for i := 0; i < mapCount; i++ {
	// 	maps = append(maps, om.New())
	// 	maps[i].ReplicateTo(tgt)
	// 	maps[i].Set(maps[i], "akey", "aval")
	// }
	/*	m := om.New()
		log.Println(m)
		m.Set("me", "key", "mval")
		m.Set("you", "key", "yval")
		m.Set("me", "2key2", "val")
		log.Println("m", m)

		repl1 := om.New()
		m.ReplicateTo(repl1)
		repl2 := om.New()
		m.ReplicateTo(repl2)
		log.Println("m", m)
		log.Println("repl1", repl1)
		m.Set("them", "bar", "baz")
		m.DeleteOwner("you")
	*/
	// time.Sleep(time.Second)
	// log.Println(tgt)
	go initWeb()
	cluster.RegisterMsg(Ping{})
	cluster.RegisterMsg(Pong{})
	log.Fatal(cluster.Init())
	/*	log.Println("m", m)
		log.Println("repl1", repl1)
		log.Println("repl2", repl2)
	*/
}

type Ping struct {
}
type Pong struct {
}

/*
func main2() {
	flag.Parse()
	testMap()
	startNode()
}

type DisGommand struct {
	Action string
	Key    string
	Value  string
}

type DisGommandable interface {
	ToDisGommands(action string) DisGommands
}

type DisGommands []DisGommand

func registryLoop() {
	reg := make(map[string]*NodeConnection)
	for {
		select {
		case name := <-unregister:
			nc, ok := reg[name]
			if ok {
				log.Println("Unregistered:", nc.node.Name)
				delete(reg, name)
			} else {
				log.Println("Unknown node:", name)
			}

		case nc := <-register:
			cur, ok := reg[nc.node.Name]
			if ok {
				log.Printf("Can't register: %s, due to conflict with: %s", nc.node, cur.node)
				nc.Close()
			} else {
				log.Printf("Registered: %s", nc.node.Name)
				f := nc.node.ToDisGommands("put")
				for _, peer := range reg {
					err := peer.Send(f)
					if err != nil {
						log.Println("Error sending:", err)
					} else {
						log.Printf("sent %+v -- %+v", peer.node, f)
					}
				}
				// nc.out.Encode(&nodeFound)
				reg[nc.node.Name] = nc
				go nodeLoop(nc)
			}
		}
	}
}

func nodeList(reg map[string]*NodeConnection) []*Node {
	fmt.Println(len(reg), reg)
	ret := make([]*Node, 0, len(reg))
	for _, n := range reg {
		ret = append(ret, n.node)
	}
	return ret
}

func nodeLoop(node *NodeConnection) {
	for {
		var msg DisGommands
		log.Println("Waiting for msg")
		err := node.in.Decode(&msg)
		log.Println("GOT stuff")
		if err != nil {
			if err == io.EOF {
				log.Printf("Disconnected from: %+v", node)
			} else {
				log.Printf("Unexepcted Error: %s, from: %s", err, node)
			}
			node.conn.Close()
			unregister <- node.node.Name
			return
		} else {
			log.Printf("Got msg: %#v", msg)
		}
	}
}
func initConnection(conn net.Conn) error {
	log.Println("Intitializing connection with:", conn.RemoteAddr())

	enc := gob.NewEncoder(conn)
	dec := gob.NewDecoder(conn)

	err := enc.Encode(&nodeConfig.MyNode)
	var theirHello Node

	if err == nil {
		err = dec.Decode(&theirHello)
	}
	if err != nil {
		conn.Close()
		log.Println("Error during handshake with", conn.RemoteAddr(), err)
		return err
	}
	if nodeConfig.MyNode.Name == theirHello.Name {
		if nodeConfig.MyNode.BootNanos < theirHello.BootNanos {
			log.Printf("Duplicate name: %s detected, other should die, their hello is: %#v", nodeConfig.MyNode.Name, theirHello)
		} else {
			log.Fatalf("Exiting due to duplicate name: %s detected, other node is: %#v", nodeConfig.MyNode.Name, theirHello)
		}
	}
	log.Printf("Received: %#v\n", theirHello)
	register <- &NodeConnection{node: &theirHello, in: dec, out: enc, conn: conn}
	return nil
}

func initListener() {
	nodeConfig.MyNode.Pid = os.Getpid()
	log.Println("Setting up:", nodeConfig.ListenAddr)
	listener, err := net.Listen("tcp", nodeConfig.ListenAddr)
	if err != nil {
		log.Fatal(err)
	}
	tcpAddr := listener.Addr().(*net.TCPAddr)
	if tcpAddr.IP.IsUnspecified() {
		host, err := os.Hostname()
		ensureNotError(err)
		nodeConfig.MyNode.Address = fmt.Sprintf("%s:%d", host, tcpAddr.Port)
	} else {
		nodeConfig.MyNode.Address = tcpAddr.String()
	}
	log.Printf("My hello is: %#v", nodeConfig.MyNode)
	go acceptLoop(listener)
}

func acceptLoop(listener net.Listener) {
	log.Println("Waiting for TCP connections on:", listener.Addr())
	for {
		con, err := listener.Accept()
		if err != nil {
			listener.Close()
			log.Fatal(err)
		}
		err = initConnection(con)
		if err != nil {
			listener.Close()
			log.Println("Error during accept handshake with", con.RemoteAddr(), err)
		}
	}
}

func attemptJoin(address string) error {
	log.Println("Joining:", address)
	con, err := net.Dial("tcp", address)
	if err != nil {
		return err
	}
	return initConnection(con)
}

func ensureNotError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

type AtomicInt int64

func (a *AtomicInt) inc() int64 {
	return atomic.AddInt64((*int64)(a), 1)
}
*/
