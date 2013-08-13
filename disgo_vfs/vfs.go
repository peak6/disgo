package disgo_vfs

import (
	"fmt"
	"strings"
	"time"
)

const (
	Added = 1 << iota
	Removed
	Modified
)

/*type cmd struct {
	reply chan string
}
type mkdir struct {
	cmd
	name string
}

func inloop(cmd chan interface{}) {
	for {
		select {
		case c <- cmd:
			select c.type(){

			}
		}
	}
}
*/
type ChangeListener chan Change

type Change struct {
	changeType int
	inode      INode
}

type INode interface {
	Name() string
	Parent() DirNode
	AddListener(listener ChangeListener)
	RemoveListener(listener ChangeListener)
	Notify(change Change)
}
type DirNode interface {
	INode
	// ChDir(name string) (DirNode, error)
	MkDir(name string) (DirNode, error)
	List() map[string]INode
	// CreateFile(name string) (FileNode, error)
	// Link(name string, INode target) (LinkNode, error)
}

type inode struct {
	name      string
	parent    DirNode
	listeners map[ChangeListener]bool
}
type dirnode struct {
	inode
	children map[string]INode
}

func mkdirs(parent DirNode, path string) (DirNode, error) {
	var cur DirNode = parent
	var err error
	for _, p := range strings.Split(path, "/")[1:] {
		cur, err = cur.MkDir(p)
		if err != nil {
			return nil, err
		}
	}
	return cur, nil
}

func (i *inode) AddListener(listener ChangeListener) {
	i.listeners[listener] = true
}

func (i *inode) RemoveListener(listener ChangeListener) {
	delete(i.listeners, listener)
}

func (i *inode) Name() string {
	return i.name
}
func (i *inode) Parent() DirNode {
	return i.parent
}
func (d *dirnode) List() map[string]INode {
	return d.children
}
func (i *inode) Notify(change Change) {
	for ch, _ := range i.listeners {
		ch <- change
	}
	if i.parent != nil {
		i.parent.Notify(change)
	}
}
func (d *dirnode) MkDir(name string) (DirNode, error) {
	cur, ok := d.children[name]
	if ok {
		return cur.(DirNode), nil
	}
	ret := NewDirNode(name, d)
	// fmt.Println("Assigning child:", name, "to:", d)
	d.children[name] = ret
	d.Notify(Change{Added, ret})
	return ret, nil
}

func newINode(name string, parent DirNode) inode {
	return inode{name, parent, make(map[ChangeListener]bool)}
}
func NewDirNode(name string, parent DirNode) DirNode {
	return &dirnode{newINode(name, parent), make(map[string]INode)}
}

func VFSTest() {
	root := NewDirNode("ROOT", nil)
	ch := make(chan Change)
	go func() {
		for change := range ch {
			fmt.Printf("CHANGE %+v\n", change.inode.Name())
		}
	}()
	root.AddListener(ch)
	fmt.Printf("Root: %+v\n", root)
	_, err := root.MkDir("foo")
	chkErr(err)
	bar, err := root.MkDir("bar")
	chkErr(err)
	bar.MkDir("baz")
	fmt.Printf("bar: %+v\n", bar)
	fmt.Println(mkdirs(root, "/foo/bar/baz/bing/bong/bozo"))
	time.Sleep(100 * time.Millisecond)
	fmt.Printf("Root: %+v\n", root)
}
func chkErr(err error) {
	if err != nil {
		panic(err)
	}

}
