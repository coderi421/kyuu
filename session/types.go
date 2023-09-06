package session

import (
	"context"
	"net/http"
)

// Session 这个通常是一个接口，用来表示一个 session 结构体必须要实现的方法
// session 对应的结构体，需要存在 Store 里面
type Session interface {
	// Get 获取 session 的值
	Get(ctx context.Context, key string) (string, error)
	// Set 设置 session 的值
	Set(ctx context.Context, key string, val string) error
	// ID 获取 session 的 ID
	ID() string
}

type Store interface {
	// Generate 生成一个 session
	Generate(ctx context.Context, id string) (Session, error)
	// Refresh 这种设计是一直用同一个 id 的
	// 如果想支持 Refresh 换 ID，那么可以重新生成一个，并移除原有的
	// 又或者 Refresh(ctx context.Context, id string) (Session, error)
	// 其中返回的是一个新的 Session
	Refresh(ctx context.Context, id string) error
	Remove(ctx context.Context, id string) error
	Get(ctx context.Context, id string) (Session, error)
}

// Propagator 处理请求中的 session id
type Propagator interface {
	// Inject 将 session id 注入到里面
	// Inject 必须是幂等的
	Inject(id string, writer http.ResponseWriter) error
	// Extract 将 session id 从 http.Request 中提取出来
	// 例如从 cookie 中将 session id 提取出来
	Extract(req *http.Request) (string, error)
	// Remove 将 session id 从 http.ResponseWriter 中删除
	// 例如删除对应的 cookie
	Remove(writer http.ResponseWriter) error
}
