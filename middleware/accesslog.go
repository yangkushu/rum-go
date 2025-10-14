package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/yangkushu/rum-go/iface"
	"time"
)

// AccessLog 请求日志中间件
type AccessLog struct {
	log iface.ILogger
}

func NewAccessLog(log iface.ILogger) *AccessLog {
	return &AccessLog{
		log: log,
	}
}

func (a *AccessLog) HandlerFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		//query := c.Request.URL.RawQuery
		c.Next()
		end := time.Now()
		latency := end.Sub(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		comment := c.Errors.ByType(gin.ErrorTypePrivate).String()
		a.log.Info(fmt.Sprintf("| %3d | %13v | %15s | %s  %s | %s |",
			statusCode,
			latency,
			clientIP,
			method,
			path,
			comment,
		))
	}
}
