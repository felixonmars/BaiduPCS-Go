// Package requester 提供网络请求简便操作
package requester

var (
	// UserAgent 浏览器标识
	UserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_2) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.132 Safari/537.36"

	// DefaultClient 默认 http 客户端
	DefaultClient = NewHTTPClient()
)

// ContentTyper Content-Type 接口
type ContentTyper interface {
	ContentType() string
}

// ContentLengther Content-Length 接口
type ContentLengther interface {
	ContentLength() int64
}
