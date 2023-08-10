package kyuu

import (
	"fmt"
	"strings"
)

type router struct {
	trees map[string]*node
}

func newRouter() router {
	return router{
		trees: map[string]*node{},
	}
}

// addRoute register the route into tree
//
//	@Description:
//	@receiver r
//	@param method HTTP 方法
//	@param path 必须以 / 开始并且结尾不能有 /，中间也不允许有连续的 /
//	@param handleFunc 路由处理函数
func (r *router) addRoute(method string, path string, handleFunc HandleFunc) {
	// validate route before add
	r.validateRoute(path)

	// 注册路由到路由树
	root, ok := r.trees[method]
	if !ok {
		// 创建根节点
		root = &node{path: "/"}
		r.trees[method] = root
	}

	// 如果注册的是根节点，则直接处理后返回
	if path == "/" {
		if root.handler != nil {
			panic("kyuu: 路由冲突[/]")
		}
		root.handler = handleFunc
		return
	}

	// 跳过第一个 /
	segments := strings.Split(path[1:], "/")

	// 分别处理每一段 /a/b/c
	for _, s := range segments {
		if s == "" {
			panic(fmt.Sprintf("kyuu: 非法路由。不允许使用 //a/b, /a//b 之类的路由，[%s]", path))
		}

		// 每次这里的 root 都是上一次的 child
		root = root.childOrCreate(s)
	}
	// 如果 handler 不为空，则说明路由冲突
	if root.handler != nil {
		panic(fmt.Sprintf("kyuu: 路由冲突[%s]", path))
	}
	// 找到最后一个节点，将 handler 赋值
	root.handler = handleFunc
}

func (r *router) findRoute(method string, path string) (*node, bool) {

	root, ok := r.trees[method]
	if !ok {
		return nil, false
	}

	if path == "/" {
		return root, true
	}

	segments := strings.Split(strings.Trim(path, "/"), "/")
	// 分别处理每一段 /a/b/c
	for _, s := range segments {
		root, ok = root.childOf(s)
		if !ok {
			return nil, false
		}
	}
	return root, true
}

// validateRoute
//
//	@Description:
//	@receiver r
//	@param path
func (r *router) validateRoute(path string) {
	if strings.TrimSpace(path) == "" {
		panic("kyuu: 路由是空字符串")
	}
	if path[0] != '/' {
		panic("kyuu: 路由必须以 / 开头")
	}

	if path != "/" && path[len(path)-1] == '/' {
		panic("kyuu: 路由不能以 / 结尾")
	}
}

type node struct {
	// path 是当前节点的路径
	path string
	// children 是子节点
	// key 是子节点的路径
	// value 是子节点 => *node
	children map[string]*node
	handler  HandleFunc // handler 是当前节点的业务逻辑
}

// find the child node by path or create a new node
// 查找子节点，如果子节点不存在就创建一个
// 并且将子节点放回去了 children 中
func (n *node) childOrCreate(path string) *node {
	if n.children == nil {
		child := &node{path: path}
		n.children = map[string]*node{}
		n.children[path] = child
		return child
	}

	child, ok := n.children[path]
	if !ok {
		// 如果没有找到，则创建一个新的节点
		child = &node{path: path}
		n.children[path] = child
	}
	return child
}

// childOf find the child node by path
//
//	@Description:
//	@receiver n
//	@param path
//	@return child 返回对应子节点
//	@return found
func (n *node) childOf(path string) (child *node, found bool) {
	// 如果 children 为空，则直接返回 防止 nil pointer
	if n.children == nil {
		return nil, false
	}
	child, found = n.children[path]
	return
}
