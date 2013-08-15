package main

import (
	"encoding/gob"
	"net"
)


func (n *Node) ToDisGommands(action string) DisGommands {
	return DisGommands{
		DisGommand{Action: action, Key: "Name", Value: n.Name},
		DisGommand{Action: action, Key: "Env", Value: n.Env},
	}
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

type ControlMsg struct {
	Action string
	Id     int
	Nodes  []*Node
}
