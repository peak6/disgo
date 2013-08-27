package cluster

import (
	"encoding/gob"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	log_pkg "log"
	"net"
	"os"
	"sync"
	"time"
)

var clusterWait sync.WaitGroup
var mcastAddr = "239.20.0.128:4777"
var listenAddr = ":4760"
var MyNode Node
var MyHost string
var log = log_pkg.New(os.Stderr, "[cluster] ", log_pkg.LstdFlags|log_pkg.Lshortfile|log_pkg.Lmicroseconds)
var initError error
var joinAddr string

func SendAll(msg interface{}) {

}
func Send(node Node, msg interface{}) {

}

func RegisterMsg(msg interface{}) {
	gob.Register(msg)
}

type Node struct {
	Name        string
	TcpAddr     string
	Version     string
	Environment string
	BootNanos   int64
}

func Init() error {
	if initError != nil {
		return initError
	}
	lsnr, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return err
	}
	MyNode.TcpAddr = fmt.Sprintf("%s:%d", MyHost, lsnr.Addr().(*net.TCPAddr).Port)
	log.Println("Listening for cluster connections on:", listenAddr)
	if joinAddr != "" {
		log.Println("Attempting to auto-join", joinAddr)
		conn, err := net.Dial("tcp", joinAddr)
		if err != nil {
			return err
		}
		newHandler(conn)
	}
	// mcloop()
	acceptLoop(lsnr)
	clusterWait.Wait()
	log.Println("Cluster shutdown")
	return nil
}

func join(node Node) {
	_, ok := nodes[node]
	if ok {
		log.Println("Already know:", node)
		return
	}
	log.Println("Attempting to join:", node)
	conn, err := net.Dial("tcp", joinAddr)
	if err != nil {
		log.Println("Failed to join:", err, node)
		return
	}
	newHandler(conn)
}
type Opt struct {
	Foo string[string] `short:"a"`
}
func init() {
	nodes = make(map[Node]*nodeConn)
	MyHost, initError = os.Hostname()
	if initError != nil {
		return
	}
	gob.Register(MyNode)
	MyNode.BootNanos = time.Now().UnixNano()
	MyNode.TcpAddr = MyHost + ":4760"
	flag.StringVar(&listenAddr, "claddr", listenAddr, "Address to listen for incoming connections")
	flag.StringVar(&mcastAddr, "clmcast", mcastAddr, "Multicast address to use for auto-discovery")
	flag.StringVar(&MyNode.Name, "clname", MyHost, "Node name to use")
	flag.StringVar(&joinAddr, "cljoin", joinAddr, "Node to attemp to join at start")
}

func acceptLoop(listener net.Listener) error {
	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		err = newHandler(conn)
		if err != nil {
			log.Println("Failed to create new connection")
		}
	}
}

var nodes map[Node]*nodeConn

type command struct {
}

func nodeList() []Node {
	nlist := []Node{}
	for n, _ := range nodes {
		nlist = append(nlist, n)
	}
	return nlist
}

func newHandler(conn net.Conn, initiator bool) error {
	log.Println("Setting up", conn.RemoteAddr(), conn.LocalAddr())
	nc := newNodeConn(conn)
	theirNode := Node{}

	if !initiator {
		nc.send(MyNode)
	} else {
		err := nc.receive(&theirNode)
		if err != nil {
			return err
		}

	}
	nc.send(nodeList())

	err := nc.receive(&theirNode)
	if err != nil {
		return nc.shutdown(err)
	}
	theirList := []Node{}
	err = nc.receive(&theirList)
	if err != nil {
		return nc.shutdown(err)
	}

	_, ok := nodes[theirNode]
	if ok {
		return nc.shutdown(errors.New("Duplicate connection"))
	}
	log.Println("Their node:", theirNode)
	log.Println("Their node list: ", theirList)
	nc.node = theirNode
	nodes[theirNode] = &nc
	clusterWait.Add(1)
	nc.startCmdLoop()
	for _, n := range theirList {
		join(n)
	}
	return nil
}

type nodeConn struct {
	node       Node
	conn       net.Conn
	start      time.Time
	cmdChan    chan command
	enc        *gob.Encoder
	dec        *gob.Decoder
	isShutdown bool
}

func newNodeConn(conn net.Conn) nodeConn {
	return nodeConn{
		conn:    conn,
		start:   time.Now(),
		cmdChan: make(chan command),
		dec:     gob.NewDecoder(conn),
		enc:     gob.NewEncoder(conn),
	}
}
func (nc *nodeConn) send(msg interface{}) error {
	return nc.enc.Encode(msg)
}
func (nc *nodeConn) receive(msg interface{}) error {
	return nc.dec.Decode(msg)
}
func (nc *nodeConn) shutdown(err error) error {
	if nc.isShutdown {
		return err
	}
	delete(nodes, nc.node)
	nc.isShutdown = true
	nc.conn.Close()
	close(nc.cmdChan)
	log.Println("Connection:", nc.conn.RemoteAddr(),
		"closed after:", time.Now().Sub(nc.start),
		"reason:", err)
	return err
}

func (nc *nodeConn) startCmdLoop() {
	clusterWait.Add(2)
	go func() {
		defer clusterWait.Done()
		defer nc.shutdown(errors.New("Unknown reason"))
		for msg := range nc.cmdChan {
			log.Println("Sending:", msg)
			err := nc.send(msg)
			if err != nil {
				log.Panic(err)
			}
		}
	}()
	go func() {
		defer clusterWait.Done()
		defer nc.shutdown(errors.New("Unkown reason"))
		for {
			var msg command
			err := nc.receive(&msg)
			if err != nil {
				nc.shutdown(err)
				return
			}
			log.Println("Got:", msg)
		}
	}()
}

func mcloop() {
	i, err := net.InterfaceByName("lo")
	gaddr, err := net.ResolveUDPAddr("udp", mcastAddr)
	if err != nil {
		log.Println("Multicast discover disabled:", err)
		return
	}
	conn, err := net.ListenMulticastUDP("udp", i, gaddr)
	if err != nil {
		log.Println("Multicast discover disabled:", err)
		return
	}
	clusterWait.Add(2)
	go func() {
		defer clusterWait.Done()
		data, err := json.Marshal(MyNode)
		if err != nil {
			log.Panic("Fail", err)
		}
		for {
			time.Sleep(time.Second)
			conn.WriteToUDP(data, gaddr)
			log.Println("sent mcast")
		}
	}()
	go func() {
		defer clusterWait.Done()
		data := make([]byte, 2048, 2048)
		for {
			fmt.Println("Reading mcast")
			sz, addr, err := conn.ReadFromUDP(data)
			fmt.Println("Read:", sz)
			if err != nil {
				log.Println("hmmm", err)
			}
			n := Node{}
			json.Unmarshal(data[:sz], &n)
			log.Println("Got mcast: ", n, addr)

		}
	}()
}
