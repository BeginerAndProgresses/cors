package cors

import (
	"cors/utils"
	"log"
	"net/http"
	"os"
	"slices"
	"strings"
)

type Cors struct {
	// 配置
	// 包含跨域相关
	allowedHeader []string
	allowedMethod []string
	allowedOrigin []string
	credentials   bool
	exposedHeader []string
	maxAge        int
	// 如果你已经在其他中间件中处理了OPTIONS请求，
	// 则设置为true
	preflightPass bool
	// 如果配置AllowOriginFunc，
	// 则忽略AllowedOrigin
	allowOriginFunc func(r *http.Request, origin string) bool
	// 开启debug模式，会输出日志
	// 默认使用控制台输出
	logger        logger
	debug         bool
	preflightVary []string
}

type logger interface {
	Printf(string, ...interface{})
}

// NewCors 根据配置创建Cors
func NewCors(options Options) *Cors {
	if options.DeBug {
		options.logger = log.New(os.Stdout, "[cors] ", log.LstdFlags)
	}
	return &Cors{
		allowedHeader:   options.AllowedHeader,
		allowedMethod:   options.AllowedMethod,
		allowedOrigin:   options.AllowedOrigin,
		credentials:     options.Credentials,
		exposedHeader:   options.ExposedHeader,
		maxAge:          options.MaxAge,
		preflightPass:   options.PreflightPass,
		allowOriginFunc: options.AllowOriginFunc,
		logger:          options.logger,
		debug:           options.DeBug,
		preflightVary:   []string{"Origin, Access-Control-Request-Method, Access-Control-Request-Headers"},
	}
}

func (c *Cors) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" && r.Header.Get("Origin") != "" {
			c.logf("preflight request: url-%s,header-%s", r.RequestURI, r.Header)
			c.handlePreflight(w, r)
			if c.preflightPass {
				next.ServeHTTP(w, r)
			} else {
				// 204
				w.WriteHeader(http.StatusNoContent)
			}
		} else {
			c.logf("actual request: url-%s,header-%s", r.RequestURI, r.Header)
			c.handleActualRequest(w, r)
			next.ServeHTTP(w, r)
		}
	})
}

func (c *Cors) handlePreflight(w http.ResponseWriter, r *http.Request) {
	header := w.Header()
	origin := r.Header.Get("Origin")
	if r.Method != http.MethodOptions {
		c.logf("preflight aborted: %s != OPTIONS", r.Method)
		return
	}
	if vary, found := header["Vary"]; found {
		header["Vary"] = append(vary, c.preflightVary...)
	} else {
		header["Vary"] = c.preflightVary
	}
	if origin == "" {
		c.logf("preflight aborted: empty origin")
		return
	}
	if !c.isAllowedOrigin(r, origin) {
		c.logf("preflight aborted: origin '%s' not allowed", origin)
		return
	}
	if !c.isAllowedMethod(r.Header.Get("Access-Control-Request-Method")) {
		c.logf("preflight aborted: method '%s' not allowed", r.Header.Get("Access-Control-Request-Method"))
		return
	}
	reqHeaders, found := utils.GetHeaderFirst(r.Header, "Access-Control-Request-Headers")
	if found && utils.ContainerOtherAll(c.allowedHeader, reqHeaders) {
		c.logf("preflight aborted: header '%s' not allowed", reqHeaders[0])
		return
	}
	if len(c.allowedOrigin) > 0 && c.allowedOrigin[0] == "*" {
		header.Set("Access-Control-Allow-Origin", "*")
	} else if len(c.allowedOrigin) > 0 {
		header["Access-Control-Allow-Origin"] = c.allowedOrigin
	} else {
		c.logf("preflight aborted: no allowed origin")
		return
	}
	header["Access-Control-Allow-Methods"] = r.Header["Access-Control-Request-Method"]
	if found && len(reqHeaders[0]) > 0 {
		header["Access-Control-Allow-Headers"] = r.Header["Access-Control-Request-Headers"]
	}
	if c.credentials {
		header["Access-Control-Allow-Credentials"] = []string{"true"}
	}
	if c.maxAge > 0 {
		header["Access-Control-Max-Age"] = []string{utils.IntToString(c.maxAge)}
	} else {
		header["Access-Control-Max-Age"] = []string{"0"}
	}
	if c.exposedHeader != nil {
		header["Access-Control-Expose-Headers"] = r.Header["Access-Control-Expose-Headers"]
	}
	c.logf("preflight response header: %s", header)
}

func (c *Cors) handleActualRequest(w http.ResponseWriter, r *http.Request) {
	header := w.Header()
	origin := r.Header.Get("Origin")

	// 总是设置Vary
	if vary := header["Vary"]; vary == nil {
		header["Vary"] = []string{"Origin"}
	} else {
		header["Vary"] = append(vary, "Origin")
	}
	// 检查Origin要在vary之后
	// 为了避免缓存请求
	// 导致后续请求携带origin也无法访问的问题
	if origin == "" {
		c.logf("actual request aborted: empty origin")
		return
	}
	if !c.isAllowedOrigin(r, origin) {
		c.logf("actual request aborted: origin '%s' not allowed", origin)
		return
	}
	if !c.isAllowedMethod(r.Method) {
		c.logf("actual request aborted: method '%s' not allowed", r.Method)
		return
	}
	if len(c.allowedOrigin) > 0 && c.allowedOrigin[0] == "*" {
		header.Set("Access-Control-Allow-Origin", "*")
	} else if len(c.allowedOrigin) > 0 {
		header["Access-Control-Allow-Origin"] = c.allowedOrigin
	} else {
		c.logf("actual request aborted: no allowed origin")
		return
	}
	if c.credentials {
		header["Access-Control-Allow-Credentials"] = []string{"true"}
	}
	c.logf("actual response header: %s", header)
}

func (c *Cors) logf(f string, ars ...interface{}) {
	if c.logger != nil && c.debug {
		c.logger.Printf(f, ars)
	}
}

// ServeHTTP 实现http.Handler接口
func (c *Cors) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" && r.Header.Get("Origin") != "" {
		c.logf("preflight request: url-%s,header-%s", r.RequestURI, r.Header)
		c.handlePreflight(w, r)
		if c.preflightPass {
			next(w, r)
		} else {
			// 204
			w.WriteHeader(http.StatusNoContent)
		}
	} else {
		c.logf("actual request: url-%s,header-%s", r.RequestURI, r.Header)
		c.handleActualRequest(w, r)
		next(w, r)
	}
}

// Default 默认配置
// 放行所有Origin
// 放行四种请求方式
// 允许携带Cookie
// Debug 模式关闭
// 开启预检请求
func Default() *Cors {
	options := Options{
		AllowedHeader: []string{"Content-Type", "Authorization"},
		AllowedMethod: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedOrigin: []string{"*"},
		Credentials:   true,
		DeBug:         false,
		ExposedHeader: []string{"Content-Type", "Authorization"},
		MaxAge:        0,
	}
	return NewCors(options)
}

func (c *Cors) isAllowedMethod(md string) bool {
	md = strings.ToUpper(md)
	return slices.Contains(c.allowedMethod, md)
}

// isAllowedOrigin 判断Origin是否在允许的Origin列表中
func (c *Cors) isAllowedOrigin(r *http.Request, origin string) bool {
	if c.allowOriginFunc != nil {
		return c.allowOriginFunc(r, origin)
	}
	if len(c.allowedOrigin) > 0 && c.allowedOrigin[0] == "*" {
		return true
	}
	origin = strings.ToLower(origin)
	allowedOrigin := utils.HandleSlice(c.allowedOrigin, strings.ToLower)
	return slices.Contains(allowedOrigin, origin)
}

// SetLog 设置日志
func (c *Cors) SetLog(logger *log.Logger) {
	c.logger = logger
}
