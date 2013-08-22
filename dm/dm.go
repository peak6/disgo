package dm

import (
	"errors"
	"fmt"
)

func NewDM() *Elem {
	return &Elem{name: "$ROOT"}
}

type Elem struct {
	name     string
	children map[string]Elem
	parent   *Elem
	val      interface{}
}

func (e Elem) String() string {
	var str string
	if e.parent != nil {
		str = e.parent.String() + e.name
	} else {
		str = e.name
	}

	if e.val != nil {
		str = fmt.Sprintf("%s : %v", str, e.val)
	} else if e.children != nil {
		str += "/"
	}
	return str
}

func (e *Elem) Set(val interface{}) error {
	if e.children != nil {
		return errors.New("node has children, can't set value")
	}
	e.val = val
	return nil
}

func (e *Elem) AddChild(name string) (*Elem, error) {
	if e.val != nil {
		return nil, errors.New("node has a value, can't add children")
	}
	if e.children == nil {
		e.children = make(map[string]Elem)
	} else {
		_, ok := e.children[name]
		if ok {
			return nil, errors.New("child already exists")
		}
	}
	nn := Elem{parent: e, name: name}
	e.children[name] = nn
	return &nn, nil
}
