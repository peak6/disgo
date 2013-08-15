package main

import (
	"fmt"
	"sync"
	"time"
)

const (
	watchPut = iota
	watchDel
)



type dmap struct {
	lock     sync.RWMutex
	data     map[OwnerKey]interface{}
	watchers map[DMapWatcher]bool
}

func (dm *dmap) String() string {
	return fmt.Sprint(dm.data)
}

var dmaps map[string]*dmap
var dmaps_lock sync.RWMutex

type DMChangeEvent struct {
	changeType int
	dm         *dmap
	ownerKey   OwnerKey
	val        interface{}
}

type DMapWatcher chan DMChangeEvent

type OwnerKey struct {
	Owner interface{}
	Key   string
}

type addWatcher struct {
	watcher DMapWatcher
	mask    int
}
type delWatcher struct {
	watcher DMapWatcher
	mask    int
}

func init() {
	dmaps = make(map[string]*dmap)
}

func newMap(name string) *dmap {
	dmaps_lock.Lock()
	defer dmaps_lock.Unlock()
	dm := &dmap{
		data:     make(map[OwnerKey]interface{}),
		watchers: make(map[DMapWatcher]bool),
	}

	dmaps[name] = dm
	return dm
}

func getmap(name string) *dmap {
	dmaps_lock.RLock()
	ret, ok := dmaps[name]
	dmaps_lock.RUnlock()
	if ok {
		return ret
	}
	return newMap(name)
}

func (dm *dmap) AddWwatcher(dmw DMapWatcher) {
	dm.lock.Lock()
	defer dm.lock.Unlock()
	dm.watchers[dmw] = true
}
func (dm *dmap) RemWatcher(dmw DMapWatcher) {
	dm.lock.Lock()
	defer dm.lock.Unlock()
	delete(dm.watchers, dmw)
}

func (dm *dmap) getByKey(key string) map[OwnerKey]interface{} {
	dm.lock.RLock()
	defer dm.lock.RUnlock()
	ret := make(map[OwnerKey]interface{}, len(dm.data))
	for k, v := range dm.data {
		if k.Key == key {
			ret[k] = v
		}
	}
	return ret
}
func (dm *dmap) get(owner interface{}, key string) interface{} {
	dm.lock.RLock()
	defer dm.lock.RUnlock()
	return dm.data[OwnerKey{owner, key}]
}
func notify(dm *dmap, msg DMChangeEvent) {
	for w, _ := range dm.watchers {
		w <- msg
	}
}
func (dm *dmap) put(owner interface{}, key string, val interface{}) {
	ok := OwnerKey{Owner: owner, Key: key}
	dm.lock.Lock()
	defer dm.lock.Unlock()
	dm.data[ok] = val
	dce := DMChangeEvent{
		changeType: watchPut,
		ownerKey:   ok,
		dm:         dm,
		val:        val,
	}
	notify(dm, dce)
}
func tf(obj ...interface{}) {
	fmt.Println(obj...)
}
func watcher(ch DMapWatcher) {
	for msg := range ch {
		fmt.Printf("watcher:%+v\n", msg)
	}
	fmt.Println("Exiting watcher")
}

func testMap() {
	tf("getmap:foo", getmap("foo"))
	dmw := make(DMapWatcher)
	go watcher(dmw)
	dm := getmap("foo")
	dm.AddWwatcher(dmw)
	x := new(AtomicInt)
	fmt.Println(x.inc())
	fmt.Println(x.inc())
	fmt.Println(x.inc())
	dm.put(nil, "bar", "baz")
	fmt.Println("bar is", dm.get(nil, "bar"))
	fmt.Println("bing is", dm.get(nil, "bing"))
	dm.put(1, "bar", "baz2")
	fmt.Println("bar is", dm.get(nil, "bar"))
	fmt.Println("bykey bar is", dm.getByKey("bar"))
	fmt.Println("End result", getmap("foo"))
	time.Sleep(50 * time.Millisecond)
}
func (ok *OwnerKey) String() string {
	return fmt.Sprintf("{Owner:%v, Key:%s", ok.Owner, ok.Key)
}
