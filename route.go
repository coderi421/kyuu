package kyuu

import (
	"fmt"
	"regexp"
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
//		@Description:
//	  - 已经注册了的路由，无法被覆盖。例如 /user/home 注册两次，会冲突
//	  - path 必须以 / 开始并且结尾不能有 /，中间也不允许有连续的 /
//	  - 不能在同一个位置注册不同的参数路由，例如 /user/:id 和 /user/:name 冲突
//	  - 不能在同一个位置同时注册通配符路由和参数路由，例如 /user/:id 和 /user/* 冲突
//	  - 同名路径参数，在路由匹配的时候，值会被覆盖。例如 /user/:id/abc/:id，那么 /user/123/abc/456 最终 id = 456
//	    @receiver r
//	    @param method HTTP 方法
//	    @param path
//	    @param handleFunc 路由处理函数
func (r *router) addRoute(method string, path string, handleFunc HandleFunc, ms ...Middleware) {
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
		// 在这个逻辑中 可以预先 将所有和全路径路由配置的所有 middlewares 都注册到 matched middleware 中
		root = root.childOrCreate(s)
	}
	// 如果 handler 不为空，则说明路由冲突
	if root.handler != nil {
		panic(fmt.Sprintf("kyuu: 路由冲突[%s]", path))
	}
	// 找到最后一个节点，将 handler 赋值
	root.handler = handleFunc
	root.route = path
	root.mdls = ms
}

func (r *router) findRoute(method string, path string) (*matchInfo, bool) {
	root, ok := r.trees[method]
	if !ok {
		return nil, false
	}

	if path == "/" {
		return &matchInfo{n: root}, true
	}

	segments := strings.Split(strings.Trim(path, "/"), "/")
	mi := &matchInfo{}
	// 分别处理每一段 /a/b/c
	for _, s := range segments {
		var child *node

		child, ok = root.childOf(s)
		if !ok {
			// 如果没有命中任何一个，而且还有上一段路由 最后一段是 通配符 *， 那么就直接返回 通配符以后的，都归这段处理
			// /a/b/c -> 归 /a/b/*
			if root.typ == nodeTypeAny {
				mi.n = root
				return mi, true
			}
			return nil, false
		}
		// eb
		if child.paramName != "" {
			mi.addValue(child.paramName, s)
		}
		root = child
	}
	mi.n = root
	// 将这条路径上所有可能的 middlewares 一并返回
	mi.mdls = r.findMdls(root, segments)

	return mi, true
}

// findMdls find the matched routers` middlewares
func (r *router) findMdls(root *node, segs []string) []Middleware {
	queue := []*node{root}
	res := make([]Middleware, 0, 16)
	for i := 0; i < len(segs); i++ {
		seg := segs[i]
		var children []*node
		for _, cur := range queue {
			if len(cur.mdls) > 0 {
				res = append(res, cur.mdls...)
			}
			// 这里将下一层的所有可能的 子节点都找到
			children = append(children, root.childrenOf(seg)...)
		}
		// 当遍历下一段路由的时候， 将上一段收集的所有子节点赋值给队列
		queue = children
	}

	// 最后一次循环 queue 没有被执行，这里需要手动执行一下
	for _, lastCur := range queue {
		if len(lastCur.mdls) > 0 {
			res = append(res, lastCur.mdls...)
		}
	}

	return res
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

type nodeType int

const (
	// 静态路由
	nodeTypeStatic = iota
	// 正则路由
	nodeTypeReg
	// 路径参数路由
	nodeTypeParam
	// 通配符路由
	nodeTypeAny
)

// node 代表路由树的节点
// 路由树的匹配顺序是：
// 1. 静态完全匹配
// 2. 正则匹配，形式 :param_name(reg_expr)
// 3. 路径参数匹配：形式 :param_name
// 4. 通配符匹配：*
// 这是不回溯匹配
type node struct {
	typ nodeType

	// path 是当前节点的路径
	path string
	// children 子节点
	// 子节点的 path => node
	children map[string]*node
	handler  HandleFunc // handler 命中路由之后执行的逻辑
	// 注册在该节点上的 middleware
	mdls []Middleware
	// route 到达该节点的完整的路由路径
	route string

	// 通配符 * 表达的节点，任意匹配
	starChild *node

	paramChild *node
	// 正则路由和参数路由都会使用这个字段
	paramName string

	// 正则表达式
	regChild *node
	regExpr  *regexp.Regexp

	// 这个地方可能在注册路由的时候，可以为每个节点，直接注册好路由，就不用每次都找一遍了
	// 用空间换时间
	matchedMdls []Middleware
}

// childOf find the child node by path
//
//	@Description:
//	@receiver n
//	@param path
//	@return child *node 是命中的节点
//	@return found bool 代表是否命中
func (n *node) childOf(path string) (*node, bool) {
	// 如果 children 为空，则直接返回 防止 nil pointer
	if n.children == nil {
		return n.childOfNonStatic(path)
	}
	// 优先配置 静态路由
	child, ok := n.children[path]
	if !ok {
		return n.childOfNonStatic(path)
	}
	return child, ok
}

// childOfNonStatic
// 从非静态匹配的子节点里面查找 reg > param > star
//
//	@Description:
//	@receiver n
//	@param path
//	@return child *node 是命中的节点
//	@return found bool 代表是否命中
func (n *node) childOfNonStatic(path string) (*node, bool) {
	if n.regChild != nil {
		if n.regChild.regExpr.Match([]byte(path)) {
			return n.regChild, true
		}
	}
	if n.paramChild != nil {
		return n.paramChild, true
	}
	return n.starChild, n.starChild != nil
}

// find the child node by path or create a new node
// childOrCreate 查找子节点，
// 首先会判断 path 是不是通配符路径
// 其次判断 path 是不是参数路径，即以 : 开头的路径
// 最后会从 children 里面查找，
// 如果没有找到，那么会创建一个新的节点，并且保存在 node 里面
func (n *node) childOrCreate(path string) *node {
	// 通配符节点注册
	if path == "*" {
		if n.paramChild != nil {
			panic(fmt.Sprintf("kyuu: 非法路由，已有路径参数路由。不允许同时注册通配符路由和参数路由 [%s]", path))
		}
		if n.regChild != nil {
			panic(fmt.Sprintf("kyuu: 非法路由，已有正则路由。不允许同时注册通配符路由和正则路由 [%s]", path))
		}
		if n.starChild == nil {
			n.starChild = &node{path: path, typ: nodeTypeAny}
		}
		return n.starChild
	}

	// 以 : 开头，需要进一步解析，判断是参数路由还是正则路由，这里是二进制的符号
	if path[0] == ':' {
		paramName, expr, isReg := n.parseParam(path)
		if isReg {
			return n.childOrCreateReg(path, expr, paramName)
		}
		return n.childOrCreateParam(path, paramName)
	}

	if n.children == nil {
		child := &node{path: path, typ: nodeTypeStatic}
		n.children = map[string]*node{}
		n.children[path] = child
		return child
	}

	child, ok := n.children[path]
	if !ok {
		// 如果没有找到，则创建一个新的节点
		child = &node{path: path, typ: nodeTypeStatic}
		n.children[path] = child
	}
	return child
}

func (n *node) childOrCreateReg(path string, expr string, paramName string) *node {
	if n.starChild != nil {
		panic(fmt.Sprintf("kyuu: 非法路由，已有通配符路由。不允许同时注册通配符路由和正则路由 [%s]", path))
	}
	if n.paramChild != nil {
		panic(fmt.Sprintf("kyuu: 非法路由，已有路径参数路由。不允许同时注册正则路由和参数路由 [%s]", path))
	}
	if n.regChild != nil {
		if n.regChild.regExpr.String() != expr || n.paramName != paramName {
			panic(fmt.Sprintf("kyuu: 路由冲突，正则路由冲突，已有 %s，新注册 %s", n.regChild.path, path))
		}
	} else {
		regExpr, err := regexp.Compile(expr)
		if err != nil {
			panic(fmt.Errorf("kyuu: 正则表达式错误 %w", err))
		}
		n.regChild = &node{path: path, paramName: paramName, regExpr: regExpr, typ: nodeTypeReg}
	}
	return n.regChild
}

func (n *node) childOrCreateParam(path string, paramName string) *node {
	if n.regChild != nil {
		panic(fmt.Sprintf("kyuu: 非法路由，已有正则路由。不允许同时注册正则路由和参数路由 [%s]", path))
	}
	if n.starChild != nil {
		panic(fmt.Sprintf("kyuu: 非法路由，已有通配符路由。不允许同时注册通配符路由和参数路由 [%s]", path))
	}
	if n.paramChild != nil {
		if n.paramChild.path != path {
			panic(fmt.Sprintf("kyuu: 路由冲突，参数路由冲突，已有 %s，新注册 %s", n.paramChild.path, path))
		}
		return n.paramChild
	} else {
		n.paramChild = &node{path: path, paramName: paramName, typ: nodeTypeParam}
	}
	return n.paramChild
}

// parseParam
//
//	@Description:
//	@receiver n
//	@param path
//	@return string 参数名字
//	@return string 正则表达式
//	@return string 为 true 则说明是正则路由 false 则不是正则路由
func (n *node) parseParam(path string) (string, string, bool) {
	// 去除 ：
	path = path[1:]
	segments := strings.SplitN(path, "(", 2)
	// 正则参数
	if len(segments) == 2 {
		expr := segments[1]
		if strings.HasSuffix(expr, ")") {
			return segments[0], expr[:len(expr)-1], true
		}
	}

	return path, "", false
}

// 找到一个节点所有可能匹配到的节点
// /a -> 可以命中 a * 或者 a :id 或者 a regex
func (n *node) childrenOf(path string) []*node {
	// 这里 cap 为2 因为上层限定了， star param regex 不能可能同时出现 所以最多两个
	res := make([]*node, 0, 2)
	if n.starChild != nil {
		res = append(res, n.starChild)
	}
	if n.paramChild != nil {
		res = append(res, n.paramChild)
	}
	if n.regChild != nil && n.regChild.regExpr.Match([]byte(path)) {
		res = append(res, n.regChild)
	}
	if n.children != nil {
		if static, ok := n.children[path]; ok {
			res = append(res, static)
		}
	}
	return res
}

// 方便收集路径参数
type matchInfo struct {
	n          *node
	pathParams map[string]string
	mdls       []Middleware
}

func (m *matchInfo) addValue(key, value string) {
	if m.pathParams == nil {
		m.pathParams = map[string]string{key: value}
	}
	m.pathParams[key] = value
}

// type Node interface {
// 如何匹配的问题
// 在我的 web 小课
// 1. 用户是没有办法注册一个自定义的实现
// 2. 难以解决优先级的问题
// Match() bool
// }

// type staticNode struct {
//
// }
