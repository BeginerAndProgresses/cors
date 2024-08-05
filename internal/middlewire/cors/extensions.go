package cors

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"net/http"
)

/*
 * 说明：
 * 作者：吕元龙
 * 时间 2024/8/5 11:47
 */

func (c *Cors) GinHandlerFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Writer = &myGinResponseWriter{
			ResponseWriter: ctx.Writer,
			b:              bytes.NewBuffer(nil),
		}
		c.ServeHTTP(ctx.Writer, ctx.Request, func(w http.ResponseWriter, r *http.Request) {
			ctx.Next()
		})
	}
}

type myGinResponseWriter struct {
	gin.ResponseWriter
	b *bytes.Buffer
}

func (w *myGinResponseWriter) Write(b []byte) (int, error) {
	w.b.Write(b)
	return w.ResponseWriter.Write(b)
}
