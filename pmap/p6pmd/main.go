package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"sync"
	// "io"
	"log"
	"net"
	"os"
	"regexp"
	"strings"
)

var ListenAddr string
var HostName string
var registry map[string]regEntry
var regLock sync.RWMutex

type regEntry struct {
	Name     string
	Port     int
	Pid      int
	FullName string
}

func init() {
	var err error
	flag.StringVar(&ListenAddr, "l", "localhost:4370", "Address to listen on, you probably don't want to change this")
	HostName, err = os.Hostname()
	if err != nil {
		log.Fatal(err)
	}
	registry = make(map[string]regEntry)
}

func main() {
	flag.Parse()
	log.Println("Waiting for connections on:", ListenAddr)
	lsnr, err := net.Listen("tcp", ListenAddr)
	if err != nil {
		log.Fatal(err)
	}
	for {
		sock, err := lsnr.Accept()
		if err != nil {
			log.Fatalf("Receive error attempting to accept a connection: %s", err)
		}
		go hc2(NewAppConn(sock))
	}
}

func _register(name string, port int, pid int) (regEntry, error) {
	regLock.Lock()
	defer regLock.Unlock()
	fullName := name + "@" + HostName
	_, ok := registry[fullName]
	if ok {
		return regEntry{}, fmt.Errorf("Already registered: %s", name)
	}
	reg := regEntry{Name: name, Port: port, Pid: pid, FullName: fullName}
	registry[fullName] = reg
	return reg, nil
}

type errResp struct {
	Err string `json:"error"`
}

func register(ac *appConn, cmd RegCmd) bool {
	regLock.Lock()
	defer regLock.Unlock()
	if cmd.Name == "" {
		return ac.err("Name cannot be empty")
	}
	if cmd.Port == 0 {
		return ac.err("port must be > 0")
	}
	name := strings.ToLower(cmd.Name)
	port := cmd.Port
	pid := cmd.Pid
	_, ok := registry[name]
	if ok {
		ac.err(fmt.Printf("Name: %s already registered", cmd.Name))
		return false
	}
	reg := regEntry{Name: name, Port: port, Pid: pid}
	registry[name] = reg
	ac.send(Rsp{Register: &TRUE})
	return true
}

func (ac *appConn) close() {
	ac.conn.Close()
}

func (ac *appConn) send(msg interface{}) {
	ac.out.Encode(msg)
}
func (ac *appConn) err(args ...interface{}) bool {
	log.Println("Error", fmt.Sprint(args...))
	ac.send(&errResp{fmt.Sprint(args...)})
	return false
}

func NewAppConn(conn net.Conn) *appConn {
	ret := appConn{}
	ret.conn = conn
	ret.in = json.NewDecoder(conn)
	ret.out = json.NewEncoder(conn)
	ret.scanner = bufio.NewScanner(conn)
	return &ret
}

type appConn struct {
	conn    net.Conn
	in      *json.Decoder
	out     *json.Encoder
	scanner *bufio.Scanner
}
type RegCmd struct {
	Name     string
	Host     string
	Port     int
	Pid      int
	FullName string
}
type UnregCmd struct {
	Name string
}
type PingCmd struct {
}
type Cmd struct {
	Ping       *bool     `json:"ping"`
	Register   *RegCmd   `json:"register"`
	Unregister *UnregCmd `json:"unregister"`
	List       *string   `json:"list"`
}
type Rsp struct {
	Pong     *bool       `json:"pong,omitempty"`
	Register *bool       `json:"register,omitempty"`
	List     *[]regEntry `json:"list,omitempty"`
}

// type Command struct {
// 	Cmd  string `json:"cmd"`
// 	Name string `json:"name"`
// 	Port int    `json:"port"`
// 	Pid  int    `json:"pid"`
// }

func doList(ac *appConn, filter string) {
	regLock.RLock()
	defer regLock.RUnlock()
	ret := []regEntry{}
	if filter == "" || filter == "*" {
		filter = ".*"
	}
	reg, err := regexp.Compile(filter)
	if err != nil {
		log.Fatal(err)
	}
	for _, r := range registry {
		if reg.MatchString(r.Name) {
			ret = append(ret, r)
		}
	}
	ac.send(Rsp{List: &ret})
}

var TRUE = true
var FALSE = false

func hc2(ac *appConn) {
	defer ac.close()
	for {
		var cmd Cmd
		err := ac.in.Decode(&cmd)
		if err != nil {
			ac.err(err)
			return
		} else {
			if cmd.Ping != nil {
				ac.send(Rsp{Pong: &TRUE})
			} else if cmd.List != nil {
				doList(ac, *cmd.List)
			} else if cmd.Register != nil {
				if !register(ac, *cmd.Register) {
					return
				}
				log.Println("Registered:", cmd.Register)
			} else if cmd.Unregister != nil {
				log.Println("unreg")
			} else {
				log.Println("err")
			}
		}
	}
	fmt.Println("Exiting")
}

/*func handleConn(ac *appConn) {
	defer ac.close()
	// var me regEntry
	in := ac.scanner
	for in.Scan() {
		line := strings.ToLower(in.Text())
		args := strings.SplitN(line, " ", 2)
		switch args[0] {
		default:
			ac.err("Unrecognized Command:", line)
			return
		case "list":
			doList(ac, ".*")
		case "app":
			log.Println("doing app:", args[1])
			var name string
			var port int
			var pid int
			_, err := fmt.Sscan(args[1], &name, &port, &pid)
			if err != nil {
				pid = 0
				_, err = fmt.Sscan(args[1], &name, &port)
			}
			if err != nil {
				ac.err("Usage: app NAME(string) PORT(int) [PID(int)]")
				return
			}
			reg, err := register(name, port, pid)
			if err != nil {
				ac.err(err)
				return
			}
			ac.ok(reg.FullName)
		}
	}
	if err := in.Err(); err != nil {
		log.Println("Error reading connection", err)
	}
	/*	var app string
		var port int
		_, err := fmt.Fscanln(conn, &app, &port)
		if err != nil {
			log.Println("Failed to intiialize connection:", err)
			return
		}
		node := app + "@" + HostName
		log.Println("Registered:", node, "on port:", port)
		for {
			var line string
			sz, err := fmt.Fscanln(conn, &line)
			if err != nil {
				if err == io.EOF {
					log.Println("Disconnected:", node)
					return
				}
				log.Println("Error reading socket, closing:", err)
				return
			}
			log.Println("Read:", sz, "bytes, line:", line)
		}
	/
}
*/
