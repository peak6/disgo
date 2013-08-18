package main

import (
	"flag"
	"github.com/golang/glog"
	"github.com/peak6/disgo/pmap"
)

/*var ListenAddr string
var registry map[string]pmap.Registration
var regLock sync.RWMutex
var TRUE = true
var FALSE = false

func init() {
	registry = make(map[string]pmap.Registration)
}
*/
func main() {
	// Messing with google's glog module, I want it to default to log to stderr
	flag.Parse()
	if flag.Lookup("log_dir").Value.String() == "" {
		flag.Lookup("logtostderr").DefValue = "true"
		flag.Lookup("logtostderr").Value.Set("true")
	}
	defer glog.Flush()
	// log.SetFlags(log.Lmicroseconds | log.Lshortfile | log.LstdFlags)
	pmap.Server.ListenAndServe()

	glog.Info("Server exited")

}
