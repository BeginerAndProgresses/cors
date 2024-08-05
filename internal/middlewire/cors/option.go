package cors

import "net/http"

type Options struct {
	AllowedHeader []string
	AllowedMethod []string
	AllowedOrigin []string
	Credentials   bool
	ExposedHeader []string
	MaxAge        int
	// 如果你已经在其他中间件中处理了OPTIONS请求，
	// 则设置为true
	PreflightPass bool
	// 如果配置AllowOriginFunc，
	// 则忽略AllowedOrigin
	AllowOriginFunc func(r *http.Request, origin string) bool
	Logger          logger
	DeBug           bool
}
