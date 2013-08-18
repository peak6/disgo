package main

import (
	"flag"
	"fmt"
	"github.com/peak6/disgo/pmap"
	"time"
)

func main() {
	flag.Parse()
	fmt.Println("Starting test")
	client, err := pmap.NewClient()
	if err != nil {
		panic(err)
	}
	fmt.Println("client", client)
	a, b := client.Register("foo", "fooaddr")
	fmt.Println(a, b)
	a, b = client.Register("foo", "baraddr")
	fmt.Println(a, b)
	time.Sleep(5 * time.Second)
}
