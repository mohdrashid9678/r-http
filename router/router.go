package router

import (
	"strings"
	"sync"

	"github.com/mohdrashid9678/rhttp/request"
	"github.com/mohdrashid9678/rhttp/response"
)

type Handler func(*request.Request) (*response.Response, error)

// node represents a single node in the radix tree.
type node struct {
	path     string
	part     string
	children []*node
	handlers map[string]Handler // Uses the local Handler type
	isParam  bool
}

// Thread safe router type
type Router struct {
	trees map[string]*node
	mu    sync.RWMutex
}

// New creates a new Router.
func New() *Router {
	return &Router{trees: make(map[string]*node)}
}

// AddRoute now uses the local Handler type.
func (r *Router) AddRoute(method, path string, handler Handler) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.trees[method] == nil {
		r.trees[method] = &node{path: "/", part: "/"}
	}
	r.trees[method].insert(path, handler, method)
}

// FindHandler now returns the local Handler type.
func (r *Router) FindHandler(method, path string) (Handler, map[string]string) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if root := r.trees[method]; root != nil {
		return root.search(path)
	}
	return nil, nil
}

// insert adds a new route to the node's subtree.
func (n *node) insert(path string, handler Handler, method string) {
	parts := strings.Split(path, "/")[1:]
	for i, part := range parts {
		if part == "" && i == len(parts)-1 {
			break
		}
		child := n.findOrCreateChild(part)
		n = child
	}
	if n.handlers == nil {
		n.handlers = make(map[string]Handler)
	}
	n.handlers[method] = handler
}

// findOrCreateChild finds a child node for a part or creates it.
func (n *node) findOrCreateChild(part string) *node {
	for _, child := range n.children {
		if child.part == part {
			return child
		}
	}
	newChild := &node{
		part:    part,
		isParam: len(part) > 0 && part[0] == ':',
	}
	n.children = append(n.children, newChild)
	return newChild
}

// search finds a handler in the node's subtree.
func (n *node) search(path string) (Handler, map[string]string) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	params := make(map[string]string)
	currentNode := n

	for _, part := range parts {
		if part == "" {
			continue
		}
		var found bool
		for _, child := range currentNode.children {
			if child.isParam {
				params[child.part[1:]] = part
				currentNode = child
				found = true
				break
			}
			if child.part == part {
				currentNode = child
				found = true
				break
			}
		}
		if !found {
			return nil, nil
		}
	}

	if len(currentNode.handlers) > 0 {
		for _, handler := range currentNode.handlers {
			return handler, params
		}
	}
	return nil, nil
}
