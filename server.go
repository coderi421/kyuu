package kyuu

import (
	"log"
	"net"
	"net/http"
)

type HandleFunc func(ctx *Context)

// 确保 HTTPServer 肯定实现了 Server 接口
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
	// HandleFunc 是你的业务逻辑
	addRoute(method string, path string, handleFunc HandleFunc)

	// 我们并不采取这种设计方案
	// 因为后续的中断的行为 很难控制 但是在用户层面可以方便的控制
	// addRoute(method string, path string, handlers... HandleFunc)
}

// HTTPServer is the implementation of Server.
type HTTPServer struct {
	// addr string 创建的时候传递，而不是 Start 接收。这个都是可以的
	router
	mdls []Middleware
}

func NewHTTPServer() *HTTPServer {
	return &HTTPServer{
		router: newRouter(),
	}
}

// Use 可以通过调用方法注册 Middleware 也可以改成 Opts 函数选项模式
func (s *HTTPServer) Use(mdls ...Middleware) {
	if s.mdls == nil {
		s.mdls = mdls
		return
	}
	s.mdls = append(s.mdls, mdls...)
}

//// UseV1 会执行路由匹配，只有匹配上了的 mdls 才会生效
//// 这个只需要稍微改造一下路由树就可以实现
//func (s *HTTPServer) UseV1(path string, mdls ...Middleware) {
//	panic("implement me")
//}

// ServeHTTP is the entry point for a request handler.
// ServeHTTP 处理请求的入口
func (s *HTTPServer) ServeHTTP(writer http.ResponseWriter, request *http.Request) {

	// 你的框架代码就在这里
	// 将 http.Request 和 http.ResponseWriter 封装到 Context 里面
	ctx := &Context{
		Req:  request,
		Resp: writer,
	}

	// Middleware 和 serve 一起的时候， HTTPServer 执行路由匹配，应该是最后一个，最后一个执行用户的逻辑
	root := s.serve
	// 将中间件的逻辑，从后往前 将 root 放在最后一个，注册进去
	for i := len(s.mdls) - 1; i >= 0; i-- {
		root = s.mdls[i](root)
	}

	// 这里执行的时候，就是从前往后了

	// 第一个应该是回写响应的 但是注册进去的时候并未执行
	// 因为它在调用next之后才回写响应，
	// 所以实际上 flashResp 是最后一个步骤
	var m Middleware = func(next HandleFunc) HandleFunc {
		return func(ctx *Context) {
			next(ctx)
			// 用户逻辑执行完之后，进行 response 的拼凑，响应
			s.flashResp(ctx)
		}
	}
	root = m(root)
	root(ctx)
}

// serve is the core func to find the route and execute the business logic.
func (s *HTTPServer) serve(ctx *Context) {
	// 接下来就是查找路由，并且执行命中的业务逻辑
	mi, ok := s.findRoute(ctx.Req.Method, ctx.Req.URL.Path)
	if !ok || mi.n.handler == nil {
		ctx.Resp.WriteHeader(404)
		_, _ = ctx.Resp.Write([]byte("Not Found"))
		return
	}
	ctx.PathParams = mi.pathParams
	ctx.MatchedRoute = mi.n.route
	mi.n.handler(ctx)
}

func (s *HTTPServer) flashResp(ctx *Context) {
	if ctx.RespStatusCode > 0 {
		ctx.Resp.WriteHeader(ctx.RespStatusCode)
	}
	_, err := ctx.Resp.Write(ctx.RespData)
	if err != nil {
		log.Fatalln("回写响应失败", err)
	}
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

//// Start1 这样也可以
//func (s *HTTPServer) Start1(addr string) error {
//	return http.ListenAndServe(addr, s)
//}

func (s *HTTPServer) Get(path string, handleFunc HandleFunc) {
	s.addRoute(http.MethodGet, path, handleFunc)
}

func (s *HTTPServer) Post(path string, handleFunc HandleFunc) {
	s.addRoute(http.MethodPost, path, handleFunc)
}

func (s *HTTPServer) Put(path string, handleFunc HandleFunc) {
	s.addRoute(http.MethodPut, path, handleFunc)
}

func (s *HTTPServer) Delete(path string, handleFunc HandleFunc) {
	s.addRoute(http.MethodDelete, path, handleFunc)
}

func (s *HTTPServer) Patch(path string, handleFunc HandleFunc) {
	s.addRoute(http.MethodPatch, path, handleFunc)
}

func (s *HTTPServer) Options(path string, handleFunc HandleFunc) {
	s.addRoute(http.MethodOptions, path, handleFunc)
}
