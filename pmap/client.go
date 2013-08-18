package pmap

import (
	"github.com/golang/glog"
	"net"
)

type client struct {
}

var Client client

func NewClient() {
	conn, err := net.Dial("tcp", pmapAddr)
	if err != nil {
		glog.Fatalln(err)
	}
	go clientLoop(conn)
}

func clientLoop(conn net.Conn) {

}

