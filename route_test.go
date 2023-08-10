package kyuu

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"reflect"
	"testing"
)

func Test_router_addRoute(t *testing.T) {
	// 测试数据
	testRouter := []struct {
		method string
		path   string
	}{
		{method: http.MethodGet, path: "/"},
		{method: http.MethodGet, path: "/user"},
		{method: http.MethodGet, path: "/user/home"},
		{method: http.MethodGet, path: "/order/detail"},
		{method: http.MethodPost, path: "/order/create"},
		{method: http.MethodPost, path: "/login"},
	}

	mockHandler := func(ctx *Context) {}
	r := newRouter()
	// 注册路由
	for _, tr := range testRouter {
		r.addRoute(tr.method, tr.path, mockHandler)
	}

	// 期待结果
	wantRouter := &router{
		trees: map[string]*node{
			http.MethodGet: {
				path: "/",
				children: map[string]*node{
					"user": {
						path: "user",
						children: map[string]*node{
							"home": {path: "home", handler: mockHandler},
						},
						handler: mockHandler},
					"order": {
						path: "order",
						children: map[string]*node{
							"detail": {
								path:    "detail",
								handler: mockHandler,
							},
						},
					},
				},
				handler: mockHandler,
			},
			http.MethodPost: {
				path: "/",
				children: map[string]*node{
					"order": {
						path: "order",
						children: map[string]*node{
							"create": {
								path:    "create",
								handler: mockHandler,
							},
						},
					},
					"login": {
						path:    "login",
						handler: mockHandler,
					},
				},
			},
		},
	}

	msg, ok := wantRouter.equal(r)
	// 根据布尔值，返回测试返回结果
	assert.True(t, ok, msg)

	//	空字符串
	assert.PanicsWithValue(t, "kyuu: 路由是空字符串", func() {
		r.addRoute(http.MethodGet, "", mockHandler)
	})

	// 前导没有 /
	assert.PanicsWithValue(t, "kyuu: 路由必须以 / 开头", func() {
		r.addRoute(http.MethodGet, "a/b/c", mockHandler)
	})

	// 后缀有 /
	assert.PanicsWithValue(t, "kyuu: 路由不能以 / 结尾", func() {
		r.addRoute(http.MethodGet, "/a/b/c/", mockHandler)
	})

	// 根节点重复注册
	//r.addRoute(http.MethodGet, "/", mockHandler)
	assert.PanicsWithValue(t, "kyuu: 路由冲突[/]", func() {
		r.addRoute(http.MethodGet, "/", mockHandler)
	})
	// 普通节点重复注册
	r.addRoute(http.MethodGet, "/a/b/c", mockHandler)
	assert.PanicsWithValue(t, "kyuu: 路由冲突[/a/b/c]", func() {
		r.addRoute(http.MethodGet, "/a/b/c", mockHandler)
	})

	// 多个 /
	assert.PanicsWithValue(t, "kyuu: 非法路由。不允许使用 //a/b, /a//b 之类的路由，[/a//b]", func() {
		r.addRoute(http.MethodGet, "/a//b", mockHandler)
	})
	assert.PanicsWithValue(t, "kyuu: 非法路由。不允许使用 //a/b, /a//b 之类的路由，[//a/b]", func() {
		r.addRoute(http.MethodGet, "//a/b", mockHandler)
	})
}

// equal 对比 router 节点
//
//	@Description:
//	@receiver r 注册的路由
//	@param t 目标路由
//	@return string 错误信息
//	@return bool 是否相等
func (r router) equal(y router) (string, bool) {
	for k, v := range r.trees {
		yv, ok := y.trees[k]
		if !ok {
			return fmt.Sprintf("目标 router 中没有方法 %s", k), false
		}
		//	如果第一层 的key 存在，那么接下来继续对比 node 节点
		errMsg, ok := v.equal(yv)
		if !ok {
			return k + "-" + errMsg, false
		}
	}
	return "", true
}

// equal 递归对比节点是否相等
//
//	@Description:
//	@receiver n 注册的节点
//	@param t 目标节点
//	@return string 错误信息
//	@return bool 是否相等
func (n *node) equal(y *node) (string, bool) {
	if y == nil {
		return "目标节点为 nil", false
	}

	//	开始对比 path
	if n.path != y.path {
		return fmt.Sprintf("%s 节点 path 不相等 x %s, y %s", n.path, n.path, y.path), false
	}

	//	通过反射的方法对比 handleFunc 函数
	nhv := reflect.ValueOf(n.handler)
	yhv := reflect.ValueOf(y.handler)
	if nhv != yhv {
		return fmt.Sprintf("%s 节点 handler 不相等 x %s, y %s", n.path, nhv.Type().String(), yhv.Type().String()), false
	}

	//	对比长度
	if len(n.children) != len(y.children) {
		return fmt.Sprintf("%s 子节点长度不等", n.path), false
	}

	// 如果子节点长度为0 说明是 末尾 节点，直接成功
	if len(n.children) == 0 {
		return "", true
	}

	// 循环遍历 子节点 判断是否相等
	for k, v := range n.children {
		yv, ok := y.children[k]
		if !ok {
			return fmt.Sprintf("%s 目标节点，缺少子节点 %s", n.path, k), false
		}

		// 子节点进行递归调用
		errMsg, ok := v.equal(yv)
		if !ok {
			return n.path + "-" + errMsg, false
		}
	}
	return "", true
}

func Test_router_findRoute(t *testing.T) {
	testRoutes := []struct {
		method string
		path   string
	}{
		{
			method: http.MethodGet,
			path:   "/",
		},
		{
			method: http.MethodGet,
			path:   "/user",
		},
		{
			method: http.MethodPost,
			path:   "/order/create",
		},
	}

	mockHandler := func(ctx *Context) {}

	testCases := []struct {
		name     string
		method   string
		path     string
		found    bool
		wantNode *node
	}{
		{name: "method not found", method: http.MethodHead},
		{name: "path not found", method: http.MethodGet, path: "/abc"},
		{name: "root", method: http.MethodGet, path: "/", found: true, wantNode: &node{
			path:    "/",
			handler: mockHandler,
		}},
		{name: "user", method: http.MethodGet, path: "/user", found: true, wantNode: &node{
			path:    "user",
			handler: mockHandler,
		}},
		{name: "no handler", method: http.MethodPost, path: "/order", found: true, wantNode: &node{
			path: "order",
		}},
		{name: "two layer", method: http.MethodPost, path: "/order/create", found: true, wantNode: &node{
			path:    "create",
			handler: mockHandler,
		}},
	}

	r := newRouter()
	for _, tr := range testRoutes {
		r.addRoute(tr.method, tr.path, mockHandler)
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			n, found := r.findRoute(tc.method, tc.path)
			assert.Equal(t, tc.found, found)
			if !found {
				// 如果期待结果是 false 则不进行 handle 的比较
				return
			}

			// 继续对比 handle 是否相等
			wantVal := reflect.ValueOf(tc.wantNode.handler)
			nVal := reflect.ValueOf(n.handler)
			assert.Equal(t, wantVal, nVal)
		})
	}
}
