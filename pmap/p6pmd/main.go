package main

import (
	"flag"
	"github.com/peak6/disgo/pmap"
	"log"
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
	// log.SetFlags(log.Lmicroseconds | log.Lshortfile | log.LstdFlags)
	pmap.Server.ListenAndServe()

	log.Println("Server exited")

}
