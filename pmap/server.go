package pmap

import (
	"io"
	"log"
	"net"
	"regexp"
	"strings"
	"sync"
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
	defer Server.wait.Done()
	Server.registry = make(map[string]Registration)

	log.Printf("Listening for pmap clients on '%s'", pmapAddr)
	lsnr, err := net.Listen("tcp", pmapAddr)
	if err != nil {
		log.Fatal(err)
	}
	Server.wait.Add(1)
	go acceptLoop(lsnr)
	lsnr, err = net.Listen("unix", "/tmp/dlb")
	if err != nil {
		log.Fatal(err)
	}
	Server.wait.Add(1)
	go acceptLoop(lsnr)

	Server.wait.Wait()
	// for {
	// 	sock, err := lsnr.Accept()
	// 	if err != nil {
	// 		log.Fatalf("Receive error attempting to accept a connection: %s", err)
	// 	}
	// 	go connectionLoop(sock)
	// }
}

func acceptLoop(listener net.Listener) {
	defer listener.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		} else {
			go connectionLoop(conn)
		}

	}
}

func connectionLoop(conn net.Conn) {
	ac := newPMapConn(conn)
	defer ac.shutdown()
	// ac.log("Connected")
	run := true
	for run {
		var req Request
		var resp Response
		err := ac.in.Decode(&req)
		if err != nil {
			if err == io.EOF {
				break
			}
			resp.SetError(err)
			run = false
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
	}
}

func handleRegister(ac *pmapConn, reg Registration, resp *Response) {
	err := reg.Validate()
	if err != nil {
		resp.SetError(err.Error())
		return
	}
	Server.regLock.Lock()
	defer Server.regLock.Unlock()

	reg.Name = strings.ToLower(reg.Name)
	reg.SetRef(ac)
	orig, ok := Server.registry[reg.Name]
	if ok {
		resp.SetError("Already registered:", orig)
	} else {
		Server.registry[reg.Name] = reg
		resp.Register = &TRUE
		ac.log("Registered", reg)
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
