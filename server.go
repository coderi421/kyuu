package kyuu

import (
	"net"
	"net/http"
)

type handleFunc func(ctx *Context)

var _ Server = (*HTTPServer)(nil)

// Server is the interface that wraps the basic ServeHTTP method.
// here is the interface of core API
type Server interface {
	// Handler ServeHTTP should write reply headers and data to the ResponseWriter
	// 继承 http.Handler 接口
	http.Handler

	// Start starts the HTTP server.
	// addr 是监听地址。如果只指定端口，可以使用 ":8081"
	// 或者 "localhost:8082"
	Start(addr string) error

	// AddRoute add the route to the server.
	// AddRoute 路由注册功能
	// method 是 HTTP 方法
	// path 是路由
	// handleFunc 是你的业务逻辑
	AddRoute(method string, path string, handleFunc handleFunc)

	// 我们并不采取这种设计方案
	// 因为后续的中断的行为 很难控制 但是在用户层面可以方便的控制
	// addRoute(method string, path string, handlers... HandleFunc)
}

type HTTPServer struct {
	// addr string 创建的时候传递，而不是 Start 接收。这个都是可以的
}

// ServeHTTP is the entry point for a request handler.
// ServeHTTP 处理请求的入口
func (s *HTTPServer) ServeHTTP(writer http.ResponseWriter, request *http.Request) {

	// 你的框架代码就在这里
	// 将 http.Request 和 http.ResponseWriter 封装到 Context 里面
	ctx := &Context{
		Req:  request,
		Resp: writer,
	}

	s.serve(ctx)
}

// serve is the core func to find the route and execute the business logic.
func (s *HTTPServer) serve(ctx *Context) {
	// 接下来就是查找路由，并且执行命中的业务逻辑
}

// Start starts the HTTP server.
func (s *HTTPServer) Start(addr string) error {
	// 也可以自己创建 Server
	// http.Server{}
	//a :=http.Server{
	//	Addr:              "",
	//	Handler:           nil,
	//	TLSConfig:         nil,
	//	ReadTimeout:       0,
	//	ReadHeaderTimeout: 0,
	//	WriteTimeout:      0,
	//	IdleTimeout:       0,
	//	MaxHeaderBytes:    0,
	//	TLSNextProto:      nil,
	//	ConnState:         nil,
	//	ErrorLog:          nil,
	//	BaseContext:       nil,
	//	ConnContext:       nil,
	//}

	l, err := net.Listen("tcp", addr)

	if err != nil {
		return err
	}
	// 在这里，可以让用户注册所谓的 after start 回调 Hook
	// 比如说往你的 admin 注册一下自己这个实例
	// 在这里执行一些你业务所需的前置条件

	// 就是因为这里需要 addr 和 http.Handler 才能启动，所以才需要继承 http.Handler 接口
	return http.Serve(l, s)
}

// AddRoute register the route into tree
func (s *HTTPServer) AddRoute(method string, path string, handleFunc handleFunc) {
	//TODO implement me
	// 注册路由到路由树
}

// Start1 这样也可以
func (s *HTTPServer) Start1(addr string) error {
	return http.ListenAndServe(addr, s)
}

func (s *HTTPServer) Get(path string, handleFunc handleFunc) {
	s.AddRoute(http.MethodGet, path, handleFunc)
}

func (s *HTTPServer) Post(path string, handleFunc handleFunc) {
	s.AddRoute(http.MethodPost, path, handleFunc)
}

func (s *HTTPServer) Put(path string, handleFunc handleFunc) {
	s.AddRoute(http.MethodPut, path, handleFunc)
}

func (s *HTTPServer) Delete(path string, handleFunc handleFunc) {
	s.AddRoute(http.MethodDelete, path, handleFunc)
}

func (s *HTTPServer) Patch(path string, handleFunc handleFunc) {
	s.AddRoute(http.MethodPatch, path, handleFunc)
}

func (s *HTTPServer) Options(path string, handleFunc handleFunc) {
	s.AddRoute(http.MethodOptions, path, handleFunc)
}
