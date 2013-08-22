package om

import (
	"fmt"
	"log"
)

func New() *OwnerMap {
	return &OwnerMap{data: make(map[OwnerKey]interface{}), listeners: make(map[chan *Change]bool)}
}

type OwnerMap struct {
	data      map[OwnerKey]interface{}
	listeners map[chan *Change]bool
}

type OwnerKey struct {
	Owner interface{}
	Key   string
}

type ChangeType int

var ChangeSet = ChangeType(1)
var ChangeDel = ChangeType(2)

type Change struct {
	Type  ChangeType
	Owner interface{}
	Key   string
	Value interface{}
}

func (c Change) String() string {
	return fmt.Sprintf("{type: %d, Owner: %v, Key: %s, Val: %v}", c.Type, c.Owner, c.Key, c.Value)
}

func (om *OwnerMap) ReplicateTo(dest *OwnerMap) {
	if om == dest {
		panic("Cannot make a map a replica of itself")
	}
	for ok, v := range om.data {
		dest.Set(ok.Owner, ok.Key, v)
	}
	om.WatchFunc(func(ch *Change) {
		dest.ApplyChange(ch)
	})
}

func (om *OwnerMap) Unwatch(ch chan *Change) {
	delete(om.listeners, ch)
}

func (om *OwnerMap) ApplyChange(change *Change) {
	switch change.Type {
	case ChangeSet:
		om.Set(change.Owner, change.Key, change.Value)
	case ChangeDel:
		om.DeleteOwnerKey(change.Owner, change.Key)
	default:
		log.Panicf("Unknown change type: %v", change)
	}
}

// Convenience function that just watches for changes and calls the supplied function
// There is no way to terminate this goroutine
func (om *OwnerMap) WatchFunc(f func(change *Change)) {
	// This is done outside the loop since we want the watch installed before returning from WatchFunc
	ch := om.Watch()
	go func() {
		for m := range ch {
			f(m)
		}
	}()
}
func (om *OwnerMap) Watch() chan *Change {
	ch := make(chan *Change) // Allow for stacked notifications
	om.listeners[ch] = true
	return ch
}

func (om *OwnerMap) notify(change *Change) {
	for ch, _ := range om.listeners {
		ch <- change
	}
}
func (om *OwnerMap) DeleteOwner(owner interface{}) {
	for ok, _ := range om.GetOwner(owner) {
		log.Println("Removing:", ok)
		om.delok(ok)
	}
}
func (om *OwnerMap) DeleteOwnerKey(owner interface{}, key string) {
	om.delok(OwnerKey{owner, key})
}

func (om *OwnerMap) delok(ownerkey OwnerKey) {
	delete(om.data, ownerkey)
	om.notify(&Change{Type: ChangeDel, Owner: ownerkey.Owner, Key: ownerkey.Key})
}
func (om *OwnerMap) Set(owner interface{}, key string, val interface{}) {
	k := OwnerKey{owner, key}
	om.data[k] = val
	om.notify(&Change{Type: ChangeSet, Owner: owner, Key: key, Value: val})
}

func (om *OwnerMap) GetOwner(owner interface{}) map[OwnerKey]interface{} {
	ret := make(map[OwnerKey]interface{})
	for ok, v := range om.data {
		if ok.Owner == owner {
			ret[ok] = v
		}
	}
	return ret
}

func (om *OwnerMap) GetKey(key string) map[OwnerKey]interface{} {
	ret := make(map[OwnerKey]interface{})
	for ok, v := range om.data {
		if ok.Key == key {
			ret[ok] = v
		}
	}
	return ret
}
