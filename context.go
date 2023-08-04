package kyuu

import "net/http"

// Context is the interface that wraps the basic ServeHTTP method.
type Context struct {
	// 将 http.Request 和 http.ResponseWriter 封装到 Context 里面
	// 这样就可以在业务逻辑里面使用了
	Req  *http.Request
	Resp http.ResponseWriter
}
