package main

import (
	"flag"
	"fmt"
	"github.com/peak6/disgo/cluster"
	"github.com/peak6/disgo/om"
	"net/http"
)

var webAddr = ":8000"

func init() {
	flag.StringVar(&webAddr, "w", webAddr, "Web listen address")
}

func initWeb() {
	http.HandleFunc("/", root)
	http.HandleFunc("/set", setter)
	log.Println("Starting web server on:", webAddr)
	log.Fatal(http.ListenAndServe(webAddr, nil))
}

func setter(w http.ResponseWriter, r *http.Request) {
	m := r.FormValue("map")
	k := r.FormValue("key")
	v := r.FormValue("v")
	usem, ok := maps[m]
	if !ok {
		usem = om.New()
		maps[m] = usem
	}
	usem.Set(cluster.MyNode, k, v)
}

func root(w http.ResponseWriter, r *http.Request) {
	var str string
	for name, v := range maps {
		str = fmt.Sprintf("%s%s %v\n", str, name, v)
	}
	fmt.Fprintf(w, `
		<html>
		<head>
		<title>disgo</title>
		</head>
		<body>
		<form action="/set">
		Map: <input name="map" type="text" value="mymap"/> <br/>
		Key: <input name="key" type="text" value="mykey"/><br/>
		Value: <input name="val" type="text" value="myval"/></br>
		<input type="submit" value="Set"/><br/>
		</form>
		<pre>
		%s
		</pre>
		</body>
		</html>
		`, str)
}
