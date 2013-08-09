package main

import (
	"encoding/gob"
	"fmt"
	"net"
)

type Node struct {
	Name    string
	Env     string
	Version int
	Address string
	Host    string
	Pid     int
}
type nodeTS Node

func (n *Node) ToDisGommands(action string) DisGommands {
	return DisGommands{
		DisGommand{Action: action, Key: "Name", Value: n.Name},
		DisGommand{Action: action, Key: "Env", Value: n.Env},
	}
}
func (n Node) String() string {
	return fmt.Sprintf("%+v", nodeTS(n))
}

type NodeConnection struct {
	node *Node
	conn net.Conn
	out  *gob.Encoder
	in   *gob.Decoder
}

func (nc *NodeConnection) Send(msg interface{}) error {
	return nc.out.Encode(msg)
}

func (nc *NodeConnection) Close() error {
	return nc.conn.Close()
}

type nodeConnectionTS NodeConnection

func (nc NodeConnection) String() string {
	return fmt.Sprintf("%+v", nodeConnectionTS(nc))
}

type ControlMsg struct {
	Action string
	Id     int
	Nodes  []*Node
}
