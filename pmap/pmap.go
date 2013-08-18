package pmap

import (
	"flag"
	"fmt"
	"github.com/golang/glog"
	"os"
	"strings"
)

var pmapAddr string
var HostName string

func init() {
	flag.StringVar(&pmapAddr, "p6pmd", ":4370", "Address of p6pmd")
	var err error
	HostName, err = os.Hostname()
	if err != nil {
		glog.Fatal(err)
	}
}

type Registration struct {
	Name    string      `json:"name"`
	Port    int         `json:"port"`
	Address string      `json:"address"`
	Pid     int         `json:"pid"`
	ref     interface{} // Used to attach a custom value to this registration, does NOT encoded
}

func (r Registration) String() string {
	return fmt.Sprintf("{Name: %s, Address: %s, Port: %d, Pid: %d}", r.Name, r.Address, r.Port, r.Pid)
}
func (r *Registration) SetRef(ref interface{}) {
	r.ref = ref
}
func (r *Registration) GetRef() interface{} {
	return r.ref
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
