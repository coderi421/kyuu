package kyuu

import (
	"fmt"
	lru "github.com/hashicorp/golang-lru"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// FileUploader
// @Description:
type FileUploader struct {
	FileField   string                                // FileField 对应于文件在表单中的字段名字
	DstPathFunc func(fh *multipart.FileHeader) string // DstPathFunc 用于计算目标路径
}

// Handle 返回值是路由函数，这样封装了一层，可以通过传参，继续添加功能。
// 上一种可以在返回 HandleFunc 之前可以继续检测一下传入的字段
// 这种形态和 Option 模式配合就很好
func (f *FileUploader) Handle() HandleFunc {
	// 这里可以额外做一些检测
	// if f.FileField == "" {
	// 	// 这种方案默认值我其实不是很喜欢
	// 	// 因为我们需要教会用户说，这个 file 是指什么意思
	// 	f.FileField = "file"
	// }
	return func(ctx *Context) {
		// 上传文件的逻辑在这里

		// 第一步：读到文件内容
		// 第二步：计算出目标路径
		// 第三步：保存文件
		// 第四步：返回响应

		if f.DstPathFunc == nil {
			log.Fatalln("未提供实现计算目标路径方法")
		}
		if f.FileField == "" {
			log.Fatalln("Form 中提取文件标识未提供")
		}
		src, srcHeader, err := ctx.Req.FormFile(f.FileField)
		if err != nil {
			ctx.RespStatusCode = 400
			ctx.RespData = []byte("未找到指定数据，上传失败")
			log.Fatalln(err)
			return
		}
		defer src.Close()

		// 可以尝试把 dst 上不存在的目录都全部建立起来
		//os.MkdirAll()

		// O_WRONLY 写入数据
		// O_TRUNC 如果文件本身存在，清空数据
		// O_CREATE 创建一个新的
		// f.DstPathFunc(srcHeader) -> 这种做法就是，将目标路径计算逻辑，交给用户
		dst, err := os.OpenFile(f.DstPathFunc(srcHeader), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o666)

		if err != nil {
			ctx.RespStatusCode = 500
			ctx.RespData = []byte("上传失败")
			log.Fatalln(err)
			return
		}
		defer dst.Close()

		_, err = io.CopyBuffer(dst, src, nil)
		if err != nil {
			ctx.RespStatusCode = 500
			ctx.RespData = []byte("上传失败")
			log.Fatalln(err)
			return
		}
		ctx.RespStatusCode = http.StatusOK
		ctx.RespData = []byte("上传成功")
	}
}

//// HandleFunc 这种设计方案也是可以的，但是不如上一种灵活。
//// 它可以直接用来注册路由
//func (f *FileUploader) HandleFunc(ctx *Context) {
//	src, srcHeader, err := ctx.Req.FormFile(f.FileField)
//	if err != nil {
//		ctx.RespStatusCode = 400
//		ctx.RespData = []byte("上传失败，未找到数据")
//		log.Fatalln(err)
//		return
//	}
//	defer src.Close()
//	dst, err := os.OpenFile(f.DstPathFunc(srcHeader),
//		os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o666)
//	if err != nil {
//		ctx.RespStatusCode = 500
//		ctx.RespData = []byte("上传失败")
//		log.Fatalln(err)
//		return
//	}
//	defer dst.Close()
//
//	_, err = io.CopyBuffer(dst, src, nil)
//	if err != nil {
//		ctx.RespStatusCode = 500
//		ctx.RespData = []byte("上传失败")
//		log.Fatalln(err)
//		return
//	}
//	ctx.RespData = []byte("上传成功")
//}

// FileDownloader 直接操作了 http.ResponseWriter
// 所以在 Middleware 里面将不能使用 RespData
// 因为没有赋值
type FileDownloader struct {
	Dir string
}

func (f *FileDownloader) Handle() HandleFunc {
	return func(ctx *Context) {
		// 用的是 xxx?file=xxx
		req, err := ctx.QueryValue("file").String()
		if err != nil {
			ctx.RespStatusCode = http.StatusBadRequest
			ctx.RespData = []byte("找不到目标文件")
			return
		}
		path := filepath.Join(f.Dir, filepath.Clean(req))
		// 做一个校验，防止相对路径引起攻击者下载了你的系统文件
		// 防止通过 ../../ 这种路径 获取到你的 系统文件
		// dst, err = filepath.Abs(dst)
		// if strings.Contains(dst, d.Dir) {
		//
		// }
		fn := filepath.Base(path)

		header := ctx.Resp.Header()
		header.Set("Content-Disposition", "attachment;filename="+fn)
		header.Set("Content-Description", "File Transfer")
		header.Set("Content-Type", "application/octet-stream")
		header.Set("Content-Transfer-Encoding", "binary")
		header.Set("Expires", "0")
		header.Set("Cache-Control", "must-revalidate")
		header.Set("Pragma", "public")
		// 这里直接回复了，所以没法修改 Middleware 里面将不能使用 RespData
		http.ServeFile(ctx.Resp, ctx.Req, path)
	}
}

type StaticResourceHandlerOption func(*StaticResourceHandler)

// StaticResourceHandler 静态资源处理
// 两个层面上
// 1. 大文件不缓存
// 2. 控制住了缓存的文件的数量
// 所以，最多消耗多少内存？ size(cache) * maxSize
type StaticResourceHandler struct {
	dir                     string
	pathPrefix              string
	extensionContentTypeMap map[string]string

	// 缓存静态资源的限制
	cache       *lru.Cache
	maxFileSize int
}

// fileCacheItem cache 缓存用的结构体信息
type fileCacheItem struct {
	fileName    string
	fileSize    int
	contentType string
	data        []byte
}

func NewStaticResourceHandler(dir, pathPrefix string, options ...StaticResourceHandlerOption) *StaticResourceHandler {
	res := &StaticResourceHandler{
		dir:        dir,
		pathPrefix: pathPrefix,
		extensionContentTypeMap: map[string]string{
			// 这里根据自己的需要不断添加
			"jpeg": "image/jpeg",
			"jpe":  "image/jpeg",
			"jpg":  "image/jpeg",
			"png":  "image/png",
			"pdf":  "image/pdf",
		},
	}

	for _, opt := range options {
		opt(res)
	}
	return res
}

// WithFileCache 静态文件将会被缓存
// maxFileSizeThreshold 超过这个大小的文件，就被认为是大文件，我们将不会缓存
// maxCacheFileCnt 最多缓存多少个文件
// 所以我们最多缓存 maxFileSizeThreshold * maxCacheFileCnt
func WithFileCache(maxFileSizeThreshold int, maxCacheFileCnt int) StaticResourceHandlerOption {
	return func(h *StaticResourceHandler) {
		c, err := lru.New(maxCacheFileCnt)
		if err != nil {
			log.Printf("创建缓存失败，将不会缓存静态资源")
		}
		h.maxFileSize = maxFileSizeThreshold
		h.cache = c
	}
}

func WithMoreExtension(extMap map[string]string) StaticResourceHandlerOption {
	return func(h *StaticResourceHandler) {
		for ext, contentType := range extMap {
			h.extensionContentTypeMap[ext] = contentType
		}
	}
}

func StaticWithMaxFileSize(maxSize int) StaticResourceHandlerOption {
	return func(handler *StaticResourceHandler) {
		handler.maxFileSize = maxSize
	}
}

// Handle 静态资源的处理逻辑
func (h *StaticResourceHandler) Handle(ctx *Context) {
	req, _ := ctx.PathValue("file").String()
	if item, ok := h.readFileFromData(req); ok {
		log.Printf("从缓存中读取数据...")
		h.writeItemAsResponse(item, ctx.Resp)
		return
	}
	path := filepath.Join(h.dir, req)
	f, err := os.Open(path)
	if err != nil {
		ctx.Resp.WriteHeader(http.StatusInternalServerError)
		return
	}
	ext := getFileExt(f.Name())
	t, ok := h.extensionContentTypeMap[ext]
	if !ok {
		ctx.Resp.WriteHeader(http.StatusBadRequest)
		return
	}

	data, err := ioutil.ReadAll(f)
	if err != nil {
		ctx.Resp.WriteHeader(http.StatusInternalServerError)
		return
	}
	item := &fileCacheItem{
		fileSize:    len(data),
		data:        data,
		contentType: t,
		fileName:    req,
	}

	h.cacheFile(item)
	h.writeItemAsResponse(item, ctx.Resp)
}

func (h *StaticResourceHandler) cacheFile(item *fileCacheItem) {
	if h.cache != nil && item.fileSize < h.maxFileSize {
		h.cache.Add(item.fileName, item)
	}
}

func (h *StaticResourceHandler) writeItemAsResponse(item *fileCacheItem, writer http.ResponseWriter) {
	writer.WriteHeader(http.StatusOK)
	writer.Header().Set("Content-Type", item.contentType)
	writer.Header().Set("Content-Length", fmt.Sprintf("%d", item.fileSize))
	_, _ = writer.Write(item.data)

}

func (h *StaticResourceHandler) readFileFromData(fileName string) (*fileCacheItem, bool) {
	if h.cache != nil {
		if item, ok := h.cache.Get(fileName); ok {
			return item.(*fileCacheItem), true
		}
	}
	return nil, false
}

func getFileExt(name string) string {
	index := strings.LastIndex(name, ".")
	if index == len(name)-1 {
		return ""
	}
	return name[index+1:]
}
