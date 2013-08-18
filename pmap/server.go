package pmap

import (
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	"io"
	"net"
	"regexp"
	"strings"
	"sync"
	"time"
)

var TRUE = true
var FALSE = false

var Server server

type server struct {
	startOnce sync.Once
	wait      sync.WaitGroup
	registry  map[string]Registration
	regLock   sync.RWMutex
}

func (s *server) ListenAndServe() {
	Server.wait.Add(1)
	defer Server.wait.Done()
	Server.registry = make(map[string]Registration)
	glog.Infof("Listening for pmap clients on '%s'", pmapAddr)
	lsnr, err := net.Listen("tcp", pmapAddr)
	if err != nil {
		glog.Fatal(err)
	}
	glog.Flush()
	for {
		sock, err := lsnr.Accept()
		if err != nil {
			glog.Fatalf("Receive error attempting to accept a connection: %s", err)
		}
		go connectionLoop(sock)
	}
}

type pmapConn struct {
	conn       net.Conn
	in         *json.Decoder
	out        *json.Encoder
	sockName   string
	startTime  time.Time
	exitReason string
}

func NewAppConn(conn net.Conn) *pmapConn {
	ret := pmapConn{}
	ret.conn = conn
	ret.in = json.NewDecoder(conn)
	ret.out = json.NewEncoder(conn)
	ret.sockName = conn.RemoteAddr().String()
	ret.startTime = time.Now()
	ret.exitReason = "client disconnected"
	return &ret
}

func (ac *pmapConn) shutdown() {
	Server.regLock.Lock()
	defer Server.regLock.Unlock()
	for n, r := range Server.registry {
		if r.GetRef() == ac {
			delete(Server.registry, n) // Supposadly, this is safe, delete while iterating.  You shouldn't miss any elements
			ac.log("Unregistered", n)
		}
	}
	ac.conn.Close()
	ac.log("Disconnected after:", time.Now().Sub(ac.startTime), "reason:", ac.exitReason)
}

func (ac *pmapConn) send(msg interface{}) {
	ac.out.Encode(msg)
}

func (ac *pmapConn) log(args ...interface{}) {
	glog.Infoln(append([]interface{}{"[" + ac.sockName + "]"}, args...)...)
}
func (ac *pmapConn) logf(format string, args ...interface{}) {
	glog.Infoln("[%s] %s", ac.sockName, fmt.Sprintf(format, args...))
}

func connectionLoop(conn net.Conn) {
	ac := NewAppConn(conn)
	defer ac.shutdown()
	// ac.log("Connected")
	for {
		var req Request
		var resp Response
		err := ac.in.Decode(&req)
		if err != nil {
			if err == io.EOF {
				break
			}
			resp.SetError(err)
		} else if req.Ping != nil {
			resp.Pong = &TRUE
		} else if req.List != nil {
			handleList(req.List, &resp)
		} else if req.Register != nil {
			handleRegister(ac, *req.Register, &resp)
		} else if req.Unregister != nil {
			handleUnregister(ac, *req.Unregister, &resp)
		} else {
			resp.SetError("Must request List, Register, Unregister or Ping")
		}
		ac.send(resp)
		if resp.Error != nil {
			ac.exitReason = *resp.Error
			break
		}
	}
}

func handleRegister(ac *pmapConn, reg Registration, resp *Response) {
	Server.regLock.Lock()
	defer Server.regLock.Unlock()
	if reg.Name == "" {
		resp.SetError("Name cannot be empty")
	} else if reg.Port == 0 {
		resp.SetError("port must be > 0")
	} else {
		reg.Name = strings.ToLower(reg.Name)
		reg.SetRef(ac)
		if reg.Address == "" {
			reg.Address = HostName
		}
		orig, ok := Server.registry[reg.Name]
		if ok {
			resp.SetError("Already registered:", orig)
		} else {
			Server.registry[reg.Name] = reg
			resp.Register = &TRUE
			ac.log("Registered", reg)
		}
	}
}

func handleUnregister(ac *pmapConn, name string, resp *Response) {
	Server.regLock.Lock()
	defer Server.regLock.Unlock()
	r, ok := Server.registry[name]
	if !ok || r.GetRef() != ac {
		resp.Unregister = &FALSE
	} else {
		delete(Server.registry, name)
	}
	ac.log("Unregistered", name)
	resp.Unregister = &TRUE
}

func handleList(filterPtr *string, resp *Response) {
	Server.regLock.RLock()
	defer Server.regLock.RUnlock()
	ret := []Registration{}
	filter := *filterPtr
	if filter == "" || filter == "*" {
		filter = ".*"
	}
	reg, err := regexp.Compile(filter)
	if err != nil {
		resp.SetError(err)
	}
	for _, r := range Server.registry {
		if reg.MatchString(r.Name) {
			ret = append(ret, r)
		}
	}
	resp.List = &ret
}
