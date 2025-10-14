package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/yangkushu/rum-go/log"
	"net/http/httputil"
	"net/url"
)

// 创建反向代理处理函数
func NewReverseProxy(toUrl *url.URL) gin.HandlerFunc {
	// 设置目标URL
	proxy := httputil.NewSingleHostReverseProxy(toUrl)
	log.Info("reverse proxy to: %s", toUrl)

	return func(c *gin.Context) {
		u := c.Request.URL
		m := c.Request.Method
		log.Info("reverse proxy", log.String("url", u.String()), log.String("method", m))
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}
