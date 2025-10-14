package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/yangkushu/rum-go/iface"
	"github.com/yangkushu/rum-go/log"
	"net/http"
	"strings"
)

// Cors 处理跨域中间件
type Cors struct {
	AllowedOrigins []string
	enableLog      bool
	logger         iface.ILogger
}

// OptionCors 定义配置函数类型
type OptionCors func(*Cors)

// NewCors 创建一个新的Cors中间件实例
func NewCors(allowedOrigins []string) *Cors {
	return &Cors{AllowedOrigins: allowedOrigins}
}

func NewCorsAllowAll() *Cors {
	return &Cors{AllowedOrigins: []string{}}
}

func NewCorsWithLogger(allowedOrigins []string, logger iface.ILogger) *Cors {
	return &Cors{AllowedOrigins: allowedOrigins, enableLog: true, logger: logger}
}

// HandlerFunc 返回Gin中间件处理函数
func (c *Cors) HandlerFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		origin := ctx.Request.Header.Get("Origin")
		isOriginAllowed := c.isOriginAllowed(origin)
		if c.enableLog {
			c.logger.Info("request origin", log.String("origin", origin), log.Any("AllowedOrigins", c.AllowedOrigins), log.Bool("isOriginAllowed", isOriginAllowed))
		}

		if isOriginAllowed {
			ctx.Header("Access-Control-Allow-Origin", origin)
			ctx.Header("Access-Control-Allow-Headers", "Content-Type, AccessToken, X-CSRF-Token, Authorization, Token, access-token, Psbc-Center, psbc-center")
			ctx.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, PATCH, DELETE")
			ctx.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
			ctx.Header("Access-Control-Allow-Credentials", "true")
			if ctx.Request.Method == http.MethodOptions {
				ctx.AbortWithStatus(http.StatusNoContent)
				return
			}
			ctx.Next()
			return
		}

		ctx.AbortWithStatus(http.StatusForbidden)
		return
	}
}

// isOriginAllowed 检查请求的来源是否被允许
func (c *Cors) isOriginAllowed(origin string) bool {
	// 如果 Origin 不为空，且不在允许的列表中，就返回403。否则返回200
	if len(c.AllowedOrigins) == 0 {
		return true
	}

	if origin == "" {
		return true
	}

	for _, allowedOrigin := range c.AllowedOrigins {
		if strings.EqualFold(allowedOrigin, origin) {
			return true
		}
	}

	return false
}
