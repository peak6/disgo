package main

import (
	"encoding/gob"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"
)

const (
	DISGO_VERSION = 1
)

const (
	NONE         string = "none"
	DEFAULT_NODE string = "<ENV>"
)

var waitLock sync.WaitGroup
var MyHost string
var joinTo string
var bindAddr string
var udpPort int
var tcpPort int
var MyNode Node //sent whenever a socket is established
var register chan *NodeConnection
var unregister chan string

type NotifyNode struct {
	Node *Node
}

func init() {
	hn, err := os.Hostname()
	if err != nil {
		log.Fatal(err)
	}
	gob.Register(Node{})
	gob.Register(DisGommand{})

	MyHost = hn
	MyNode.Host = MyHost
	MyNode.Version = DISGO_VERSION
	flag.StringVar(&joinTo, "j", NONE, "Join a specific node")
	flag.StringVar(&bindAddr, "b", "", "Address to bind connections")
	flag.StringVar(&MyNode.Env, "env", "dev", "Operating environment")
	flag.StringVar(&MyNode.Name, "n", DEFAULT_NODE, "Sets the node name")
	flag.IntVar(&tcpPort, "t", 8765, "TCP port to listen on")
	flag.IntVar(&udpPort, "u", 8765, "UDP port to listen on")

	register = make(chan *NodeConnection)
	unregister = make(chan string)
	go registryLoop()
}

func main() {
	flag.Parse()
	initListener()

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
				log.Println("Unregistered:", nc.node)
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
				log.Printf("Registered: %s", nc)
				f := nc.node.ToDisGommands("put")
				// f := DisGommands{
				// 	DisGommand{Action: "put", Key: "/node/" + nc.node.Address, Value: "foofie"},
				// }
				// nodeFound := &DisGommand{Action: "put", Key: "/node/" + nc.node.Address, Value: nc.node.Address}
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

	err := enc.Encode(&MyNode)
	var theirHello Node

	if err == nil {
		err = dec.Decode(&theirHello)
	}
	if err != nil {
		conn.Close()
		log.Println("Error during handshake with", conn.RemoteAddr(), err)
		return err
	}
	if MyNode.Name == theirHello.Name {
		log.Fatalf("Duplicate name: %s detected, other node is: %#v", MyNode.Name, theirHello)
	}
	log.Printf("Received: %#v\n", theirHello)
	register <- &NodeConnection{node: &theirHello, in: dec, out: enc, conn: conn}
	return nil
}

func initListener() {
	MyNode.Pid = os.Getpid()
	laddr := fmt.Sprintf("%s:%d", bindAddr, tcpPort)
	log.Println("Setting up:", laddr)
	listener, err := net.Listen("tcp", laddr)
	if err != nil {
		log.Fatal(err)
	}
	tcpAddr := listener.Addr().(*net.TCPAddr)
	if tcpAddr.IP.IsUnspecified() {
		host, err := os.Hostname()
		ensureNotError(err)
		MyNode.Address = fmt.Sprintf("%s:%d", host, tcpAddr.Port)
	} else {
		MyNode.Address = tcpAddr.String()
	}
	log.Printf("My hello is: %#v", MyNode)
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
