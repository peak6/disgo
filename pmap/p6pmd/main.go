package main

import (
	"flag"
	"fmt"
	"github.com/peak6/disgo/pmap"
	"log"
	"os"
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

var list bool
var start bool

func main() {
	flag.BoolVar(&list, "l", false, "List currently registered applications")
	flag.BoolVar(&start, "s", false, "Start portmap in foreground")
	// Messing with google's glog module, I want it to default to log to stderr
	flag.Parse()
	client, err := pmap.NewClient()

	if start {
		if err != nil {
			pmap.Server.ListenAndServe()
		} else {
			log.Fatal("Server already started")
		}
	} else {
		if err != nil {
			log.Fatalf("%s not started, error: %s", os.Args[0], err)
		}

		if list {
			items, err := client.List()
			if err != nil {
				log.Fatal(err)
			}
			if len(items) == 0 {
				fmt.Fprintln(os.Stderr, "No registered processes")
			} else {
				fmt.Printf("%-20s%-10s%s\n", "Name", "Address", "Pid")
				for _, item := range items {
					fmt.Printf("%-20s%-10s%d\n", item.Name, item.Addr, item.Pid)
				}
			}
		} else {
			log.Fatal("Error, don't know what to do")
		}
		return
	}

}
