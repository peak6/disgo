package pmap

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

var pmapAddr string
var HostName string

func init() {
	flag.StringVar(&pmapAddr, "p6pmd", ":4370", "Address of p6pmd")
	var err error
	HostName, err = os.Hostname()
	if err != nil {
		log.Fatal(err)
	}
}

type Registration struct {
	Name string      `json:"name"`
	Addr string      `json:"addr"`
	Pid  int         `json:"pid"`
	ref  interface{} // Used to attach a custom value to this registration, does NOT encoded
}

func (r Registration) String() string {
	return fmt.Sprintf("{Name: %s, Address: %s, Pid: %d}", r.Name, r.Addr, r.Pid)
}
func (r *Registration) SetRef(ref interface{}) {
	r.ref = ref
}
func (r *Registration) GetRef() interface{} {
	return r.ref
}

func (r *Registration) Validate() error {
	if r.Addr == "" {
		return errors.New("addr not specified")
	}
	if r.Name == "" {
		return errors.New("name not specified")
	}
	return nil
}

type Request struct {
	Ping       *bool         `json:"ping,omitempty"`
	Register   *Registration `json:"register,omitempty"`
	Unregister *string       `json:"unregister,omitempty"`
	List       *string       `json:"list,omitempty"`
}
type Response struct {
	Pong       *bool           `json:"pong,omitempty"`
	Register   *bool           `json:"register,omitempty"`
	Unregister *bool           `json:"unregister,omitempty"`
	List       *[]Registration `json:"list,omitempty"`
	Error      *string         `json:"error,omitempty"`
}

func (r *Response) Errorf(format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	r.Error = &s
}
func (r *Response) SetError(args ...interface{}) {
	s := strings.TrimRight(fmt.Sprintln(args...), "\n")
	r.Error = &s
}

type pmapConn struct {
	conn       net.Conn
	in         *json.Decoder
	out        *json.Encoder
	sockName   string
	startTime  time.Time
	exitReason string
	logger     *log.Logger
}

func newPMapConn(conn net.Conn) *pmapConn {
	ret := pmapConn{}
	ret.conn = conn
	ret.in = json.NewDecoder(conn)
	ret.out = json.NewEncoder(conn)
	ret.sockName = conn.RemoteAddr().String()
	ret.startTime = time.Now()
	ret.exitReason = "client disconnected"
	ret.logger = log.New(os.Stderr, "["+ret.sockName+"] ", log.LstdFlags|log.Lmicroseconds)
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
	ac.logger.Println(args...)
	// glog.Infoln(append([]interface{}{"[" + ac.sockName + "]"}, args...)...)
}
func (ac *pmapConn) logf(format string, args ...interface{}) {
	ac.logger.Printf(format, args...)
	// glog.Infoln("[%s] %s", ac.sockName, fmt.Sprintf(format, args...))
}
