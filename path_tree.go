package httpcontext

import (
	"strings"
)

type PathTreeNode struct {
	path    string
	handler ContextFunc
	filters []FilterFunc

	wildcardName string

	childrenNode []*PathTreeNode
}

func NewPathTreeNode(p string) *PathTreeNode {
	// case: if path == "", then mean url ends with '/'
	node := &PathTreeNode{
		path:         p,
		childrenNode: make([]*PathTreeNode, 0),
	}

	if name, ok := isWildcard(p); ok {
		node.wildcardName = name
	}
	return node
}

func (n *PathTreeNode) AppendChildNode(child *PathTreeNode) {
	n.childrenNode = append(n.childrenNode, child)
}

func (n *PathTreeNode) findChildNode(p string) *PathTreeNode {
	return n.matchChildNode(func(node *PathTreeNode) bool {
		return node.path == p
	})
}

func (n *PathTreeNode) matchChildNode(matcher func(*PathTreeNode) bool) *PathTreeNode {
	for _, node := range n.childrenNode {
		if matcher(node) {
			return node
		}
	}

	return nil
}

func isWildcard(path string) (string, bool) {
	l := len(path)
	if l > 3 { // {:(\w+)}
		if path[0:2] == "{:" && path[l-1] == '}' {
			return path[2 : l-1], true
		}
	}
	return "", false
}

type PathTree struct {
	root *PathTreeNode
}

func NewPathTree() *PathTree {
	return &PathTree{NewPathTreeNode("/")}
}

func (t *PathTree) Put(urlStr string, handler ContextFunc, filters ...FilterFunc) {
	if handler == nil {
		return
	}

	if urlStr == "/" { // case: special "/"
		t.root.handler = handler
		return
	}

	if urlStr[0] == '/' {
		// simple check whether the url starts with "/"
		paths := strings.Split(urlStr[1:], "/")
		t.putNode(t.root, paths, 0, handler, filters...)
	}
}

func (t *PathTree) putNode(node *PathTreeNode, paths []string, idx int, handler ContextFunc, filters ...FilterFunc) {
	if idx == len(paths) {
		return
	}

	path := paths[idx]
	curNode := node.findChildNode(path)

	if curNode == nil {
		curNode = NewPathTreeNode(path)
		node.AppendChildNode(curNode)
	}

	if idx == len(paths)-1 {
		// update handler & filters
		curNode.handler = handler
		if len(filters) > 0 {
			curNode.filters = filters
		}
	} else {
		t.putNode(curNode, paths, idx+1, handler, filters...)
	}
}

func (t *PathTree) FindHandler(urlStr string) (ContextFunc, []FilterFunc, map[string]string) {
	if urlStr == "/" {
		return t.root.handler, t.root.filters, nil
	}

	if urlStr[0] == '/' {
		paths := strings.Split(urlStr[1:], "/")
		params := map[string]string{}

		if node := t.findNode(t.root, paths, 0, params); node != nil {
			return node.handler, node.filters, params
		}
	}

	return nil, nil, nil
}

func (t *PathTree) findNode(node *PathTreeNode, paths []string, idx int, params map[string]string) *PathTreeNode {
	if idx == len(paths) {
		return nil
	}

	path := paths[idx]
	curNode := node.findChildNode(path)

	if curNode == nil {
		// find first wildcard node
		curNode = node.matchChildNode(func(cnode *PathTreeNode) bool {
			return cnode.wildcardName != ""
		})
		if curNode != nil {
			params[curNode.wildcardName] = path
		}
	}
	if curNode == nil {
		return nil
	}

	if idx == len(paths)-1 {
		return curNode
	} else {
		return t.findNode(curNode, paths, idx+1, params)
	}
}
