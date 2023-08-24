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
		// 通配符测试用例
		{method: http.MethodGet, path: "/order/*"},
		{method: http.MethodGet, path: "/*"},
		{method: http.MethodGet, path: "/*/*"},
		{method: http.MethodGet, path: "/*/abc"},
		{method: http.MethodGet, path: "/*/abc/*"},
		// 参数路由
		{method: http.MethodGet, path: "/param/:id"},
		{method: http.MethodGet, path: "/param/:id/detail"},
		{method: http.MethodGet, path: "/param/:id/*"},
		// 正则路由
		{method: http.MethodDelete, path: "/reg/:id(.*)"},
		{method: http.MethodDelete, path: "/:name(^.+$)/abc"},
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
							"home": {path: "home", handler: mockHandler, typ: nodeTypeStatic},
						},
						handler: mockHandler,
						typ:     nodeTypeStatic,
					},
					"order": {
						path: "order",
						children: map[string]*node{
							"detail": {
								path:    "detail",
								handler: mockHandler,
								typ:     nodeTypeStatic,
							},
						},
						starChild: &node{
							path:    "*",
							handler: mockHandler,
							typ:     nodeTypeAny,
						},
						typ: nodeTypeStatic,
					},
					"param": {
						path: "param",
						paramChild: &node{
							path:      ":id",
							paramName: "id",
							children: map[string]*node{
								"detail": &node{
									path:    "detail",
									handler: mockHandler,
									typ:     nodeTypeStatic,
								},
							},
							starChild: &node{
								path:    "*",
								handler: mockHandler,
								typ:     nodeTypeAny,
							},
							handler: mockHandler,
						},
					},
				},
				handler: mockHandler,
				starChild: &node{
					path: "*",
					children: map[string]*node{
						"abc": &node{
							path:      "abc",
							starChild: &node{path: "*", handler: mockHandler, typ: nodeTypeAny},
							handler:   mockHandler,
							typ:       nodeTypeStatic,
						},
					},
					handler: mockHandler,
					starChild: &node{
						path:    "*",
						handler: mockHandler,
						typ:     nodeTypeAny,
					},
					typ: nodeTypeAny,
				},
				typ: nodeTypeStatic,
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
								typ:     nodeTypeStatic,
							},
						},
						typ: nodeTypeStatic,
					},
					"login": {
						path:    "login",
						handler: mockHandler,
						typ:     nodeTypeStatic,
					},
				},
				typ: nodeTypeStatic,
			},
			http.MethodDelete: {
				path: "/",
				children: map[string]*node{
					"reg": {
						path: "reg",
						typ:  nodeTypeStatic,
						regChild: &node{
							path:      ":id(.*)",
							paramName: "id",
							typ:       nodeTypeReg,
							handler:   mockHandler,
						},
					},
				},
				regChild: &node{
					path:      ":name(^.+$)",
					paramName: "name",
					typ:       nodeTypeReg,
					children: map[string]*node{
						"abc": {
							path:    "abc",
							handler: mockHandler,
						},
					},
				},
			},
		},
	}

	msg, ok := wantRouter.equal(*r)
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

	// 同时注册通配符路由，参数路由，正则路由
	assert.PanicsWithValue(t, "kyuu: 非法路由，已有通配符路由。不允许同时注册通配符路由和参数路由 [:id]", func() {
		r.addRoute(http.MethodGet, "/a/*", mockHandler)
		r.addRoute(http.MethodGet, "/a/:id", mockHandler)
	})
	r = newRouter()
	assert.PanicsWithValue(t, "kyuu: 非法路由，已有通配符路由。不允许同时注册通配符路由和正则路由 [:id(.*)]", func() {
		r.addRoute(http.MethodGet, "/a/b/*", mockHandler)
		r.addRoute(http.MethodGet, "/a/b/:id(.*)", mockHandler)
	})
	r = newRouter()
	assert.PanicsWithValue(t, "kyuu: 非法路由，已有通配符路由。不允许同时注册通配符路由和参数路由 [:id]", func() {
		r.addRoute(http.MethodGet, "/*", mockHandler)
		r.addRoute(http.MethodGet, "/:id", mockHandler)
	})
	r = newRouter()
	assert.PanicsWithValue(t, "kyuu: 非法路由，已有路径参数路由。不允许同时注册通配符路由和参数路由 [*]", func() {
		r.addRoute(http.MethodGet, "/a/b/:id", mockHandler)
		r.addRoute(http.MethodGet, "/a/b/*", mockHandler)
	})
	r = newRouter()
	assert.PanicsWithValue(t, "kyuu: 非法路由，已有路径参数路由。不允许同时注册通配符路由和参数路由 [*]", func() {
		r.addRoute(http.MethodGet, "/:id", mockHandler)
		r.addRoute(http.MethodGet, "/*", mockHandler)
	})
	r = newRouter()
	assert.PanicsWithValue(t, "kyuu: 非法路由，已有路径参数路由。不允许同时注册正则路由和参数路由 [:id(.*)]", func() {
		r.addRoute(http.MethodGet, "/a/b/:id", mockHandler)
		r.addRoute(http.MethodGet, "/a/b/:id(.*)", mockHandler)
	})
	r = newRouter()
	assert.PanicsWithValue(t, "kyuu: 非法路由，已有正则路由。不允许同时注册通配符路由和正则路由 [*]", func() {
		r.addRoute(http.MethodGet, "/a/b/:id(.*)", mockHandler)
		r.addRoute(http.MethodGet, "/a/b/*", mockHandler)
	})
	r = newRouter()
	assert.PanicsWithValue(t, "kyuu: 非法路由，已有正则路由。不允许同时注册正则路由和参数路由 [:id]", func() {
		r.addRoute(http.MethodGet, "/a/b/:id(.*)", mockHandler)
		r.addRoute(http.MethodGet, "/a/b/:id", mockHandler)
	})
	// 参数冲突
	assert.PanicsWithValue(t, "kyuu: 路由冲突，参数路由冲突，已有 :id，新注册 :name", func() {
		r.addRoute(http.MethodGet, "/a/b/c/:id", mockHandler)
		r.addRoute(http.MethodGet, "/a/b/c/:name", mockHandler)
	})
}

// equal 对比 router 节点
//
//	@Description:
//	@receiver r 注册的路由
//	@param t 目标路由
//	@return string 错误信息
//	@return bool 是否相等
func (r *router) equal(y router) (string, bool) {
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

	// 对比类型是否对
	if n.typ != y.typ {
		return fmt.Sprintf("%s 节点类型不相等 x %d, y %d", n.path, n.typ, y.typ), false
	}

	// 对比节点参数名
	if n.paramName != y.paramName {
		return fmt.Sprintf("%s 节点参数名字不相等 x %s, y %s", n.path, n.paramName, y.paramName), false
	}

	//	对比长度
	if len(n.children) != len(y.children) {
		return fmt.Sprintf("%s 子节点长度不等", n.path), false
	}

	// 如果子节点长度为0 说明是 末尾 节点，直接成功
	if len(n.children) == 0 {
		return "", true
	}

	// 如果通配符 * 不为 nil 则继续对比
	if n.starChild != nil {
		str, ok := n.starChild.equal(y.starChild)
		if !ok {
			return fmt.Sprintf("%s 通配符节点不匹配 %s", n.path, str), false
		}
	}
	// 如果参数节点存在
	if n.paramChild != nil {
		str, ok := n.paramChild.equal(y.paramChild)
		if !ok {
			return fmt.Sprintf("%s 路径参数节点不匹配 %s", n.path, str), false
		}
	}

	// 如果正则匹配存在
	if n.regChild != nil {
		str, ok := n.regChild.equal(y.regChild)
		if !ok {
			return fmt.Sprintf("%s 路径参数节点不匹配 %s", n.path, str), false
		}
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
		{method: http.MethodGet, path: "/"},
		{method: http.MethodGet, path: "/user"},
		{method: http.MethodPost, path: "/order/create"},
		//	通配符测试
		{method: http.MethodGet, path: "/user/*/home"},
		{method: http.MethodPost, path: "/order/*"},
		// 参数路由
		{method: http.MethodGet, path: "/param/:id"},
		{method: http.MethodGet, path: "/param/:id/detail"},
		{method: http.MethodGet, path: "/param/:id/*"},
		//	正则
		{method: http.MethodDelete, path: "/reg/:id(.*)"},
		{method: http.MethodDelete, path: "/:id([0-9]+)/home"},
	}

	mockHandler := func(ctx *Context) {}

	testCases := []struct {
		name   string
		method string
		path   string
		found  bool
		mi     *matchInfo
	}{
		{name: "method not found", method: http.MethodHead},
		{name: "path not found", method: http.MethodGet, path: "/abc"},
		{name: "root", method: http.MethodGet, path: "/", found: true, mi: &matchInfo{
			n: &node{
				path:    "/",
				handler: mockHandler,
			},
		}},
		{name: "user", method: http.MethodGet, path: "/user", found: true, mi: &matchInfo{
			n: &node{
				path:    "user",
				handler: mockHandler,
			},
		}},
		{name: "no handler", method: http.MethodPost, path: "/order", found: true, mi: &matchInfo{
			n: &node{
				path: "order",
			},
		}},
		{name: "two layer", method: http.MethodPost, path: "/order/create", found: true, mi: &matchInfo{
			n: &node{
				path:    "create",
				handler: mockHandler,
			},
		}},

		//	通配符测试
		// 命中/order/*
		{name: "star match", method: http.MethodPost, path: "/order/delete", found: true, mi: &matchInfo{
			n: &node{path: "*", handler: mockHandler},
		}},
		// 命中通配符在中间的  /user/*/home
		{name: "star in middle", method: http.MethodGet, path: "/user/Tom/home", found: true, mi: &matchInfo{
			n: &node{path: "home", handler: mockHandler},
		}},
		{name: "overflow", method: http.MethodPost, path: "/order/delete/123", found: true, mi: &matchInfo{
			n: &node{path: "*", handler: mockHandler},
		}},

		// 参数路由
		// 命中 /param/:id
		{name: ":id", method: http.MethodGet, path: "/param/123", found: true, mi: &matchInfo{
			n:          &node{path: ":id", handler: mockHandler, paramName: "id"},
			pathParams: map[string]string{"id": "123"},
		}},
		// 命中 /param/:id/*
		{name: ":id/*", method: http.MethodGet, path: "/param/123/abc", found: true, mi: &matchInfo{
			n:          &node{path: "*", handler: mockHandler},
			pathParams: map[string]string{"id": "123"},
		}},
		{name: ":id/*", method: http.MethodGet, path: "/param/123/detail", found: true, mi: &matchInfo{
			n:          &node{path: "detail", handler: mockHandler},
			pathParams: map[string]string{"id": "123"},
		}},

		//	正则
		// 命中 /reg/:id(.*)
		{name: ":id(.*)", method: http.MethodDelete, path: "/reg/123", found: true, mi: &matchInfo{
			n:          &node{path: ":id(.*)", handler: mockHandler, paramName: "id"},
			pathParams: map[string]string{"id": "123"},
		}},
		// 命中 /:id([0-9]+)/home
		{name: ":id([0-9]+", method: http.MethodDelete, path: "/123/home", found: true, mi: &matchInfo{
			n:          &node{path: "home", handler: mockHandler},
			pathParams: map[string]string{"id": "123"},
		}},
		{name: "not :id([0-9]+)", method: http.MethodDelete, path: "/abc/home"},
	}

	r := newRouter()
	for _, tr := range testRoutes {
		r.addRoute(tr.method, tr.path, mockHandler)
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mi, found := r.findRoute(tc.method, tc.path)
			assert.Equal(t, tc.found, found)
			if !found {
				// 如果期待结果是 false 则不进行 handle 的比较
				return
			}
			// 对比参数是否相同
			assert.Equal(t, tc.mi.pathParams, mi.pathParams)
			// 对比节点路由
			assert.Equal(t, tc.mi.n.path, mi.n.path)

			// 继续对比 handle 是否相等
			wantVal := reflect.ValueOf(tc.mi.n.handler)
			nVal := reflect.ValueOf(mi.n.handler)
			assert.Equal(t, wantVal, nVal)
		})
	}
}
