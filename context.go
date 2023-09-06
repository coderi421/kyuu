package kyuu

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
)

// Context is the interface that wraps the basic ServeHTTP method.
type Context struct {
	// 将 http.Request 和 http.ResponseWriter 封装到 Context 里面
	// 这样就可以在业务逻辑里面使用了
	Req *http.Request
	// Resp 原生的 ResponseWriter。当你直接使用 Resp 的时候，
	// 那么相当于你绕开了 RespStatusCode 和 RespData。
	// 响应数据直接被发送到前端，其它中间件将无法修改响应
	// 其实我们也可以考虑将这个做成私有的
	Resp       http.ResponseWriter
	PathParams map[string]string
	// 缓存的响应部分
	// 这部分数据会在最后刷新
	RespStatusCode int
	RespData       []byte

	// 缓存的数据
	queryValues url.Values

	// 命中的路由
	MatchedRoute string

	// 通过 ctx 将 template engine 传递下去
	tplEngine TemplateEngine

	// 用户可以自由决定在这里存储什么，
	// 主要用于解决在不同 Middleware 之间数据传递的问题
	// 但是要注意
	// 1. UserValues 在初始状态的时候总是 nil，你需要自己手动初始化
	// 懒汉模式 => 在第一次使用的时候初始化
	UserValues map[string]any
}

func (c *Context) BindJSON(val any) error {
	if c.Req.Body == nil {
		return errors.New("kyuu: body 为 nil")
	}

	decoder := json.NewDecoder(c.Req.Body)
	// useNumber => 数字就是用 Number 来表示
	// 否则默认是 float64
	// if jsonUseNumber {
	// 	decoder.UseNumber()
	// }

	// 如果要是有一个未知的字段，就会报错
	// 比如说你 User 只有 Name 和 Email 两个字段
	// JSON 里面额外多了一个 Age 字段，那么就会报错
	// decoder.DisallowUnknownFields()
	return decoder.Decode(val)
}

func (c *Context) Render(tplName string, data any) error {
	// 不要这样子去做
	// tplName = tplName + ".gohtml"
	// tplName = tplName + c.tplPrefix
	var err error
	c.RespData, err = c.tplEngine.Render(c.Req.Context(), tplName, data)
	if err != nil {
		c.RespStatusCode = http.StatusInternalServerError
		return err
	}
	c.RespStatusCode = http.StatusOK
	return nil
}

// FormValue returns the first value for the named component of the query.
func (c *Context) FormValue(key string) StringValue {
	// ParseForm parses the raw query from the URL and updates r.Form.
	// 必须要先 ParseForm 才能使用 FormValue
	err := c.Req.ParseForm()
	if err != nil {
		return StringValue{err: err}
	}
	return StringValue{val: c.Req.FormValue(key)}
}

// QueryValue Query 和表单比起来，它没有缓存，所以需要自己缓存起来，不然每次都要解析
func (c *Context) QueryValue(key string) StringValue {
	if c.queryValues == nil {
		c.queryValues = c.Req.URL.Query()
	}

	vals, ok := c.queryValues[key]
	if !ok {
		return StringValue{err: errors.New("kyuu: 找不到这个 key")}
	}
	if len(vals) == 1 {
		return StringValue{val: vals[0]}
	}
	return StringValue{multiVal: vals}
	// 用户区别不出来是真的有值，但是值恰好是空字符串
	// 还是没有值
	// return c.queryValues.Get(key), nil
}
func (c *Context) PathValue(key string) StringValue {
	val, ok := c.PathParams[key]
	if !ok {
		return StringValue{err: errors.New("web: 找不到这个 key")}
	}
	return StringValue{val: val}
}

func (c *Context) SetCookie(cookie *http.Cookie) {
	http.SetCookie(c.Resp, cookie)
}
func (c *Context) RespJSONOK(val any) error {
	return c.RespJSON(http.StatusOK, val)
}

func (c *Context) RespJSON(code int, val any) error {
	bs, err := json.Marshal(val)
	if err != nil {
		return err
	}
	// c.Resp.WriteHeader(code)
	//	_, err = c.Resp.Write(bs)
	// 这里缓存起来，为了 after hook 处理方便
	c.RespStatusCode = code
	c.RespData = bs
	return err
}

type StringValue struct {
	val      string
	multiVal []string
	err      error
}

func (s StringValue) String() (string, error) {
	return s.val, s.err
}

func (s StringValue) StringMultiVal() ([]string, error) {
	return s.multiVal, s.err
}

func (s StringValue) ToInt64() (int64, error) {
	if s.err != nil {
		return 0, s.err
	}
	return strconv.ParseInt(s.val, 10, 64)
}
