package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/yangkushu/rum-go/iface"
	"github.com/yangkushu/rum-go/log"
	"net/http"
	"runtime"
)

// Recovery 自定义的 Recovery 中间件
type Recovery struct {
	onError func(c *gin.Context, err interface{})
	log     iface.ILogger
}

func NewRecovery(onError func(c *gin.Context, err interface{}), log iface.ILogger) *Recovery {
	return &Recovery{onError: onError, log: log}
}

func (r *Recovery) HandlerFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				if e, ok := err.(error); ok {
					r.log.Error("on recovery", log.ErrorField(e), log.String("stack", getStack()))
				} else {
					r.log.Error("on recovery", log.Any("err", err), log.String("stack", getStack()))
				}
				if r.onError != nil {
					r.onError(c, err)
				} else {
					var errMsg string
					if errStr, ok := err.(string); ok {
						errMsg = errStr
					} else {
						errMsg = "Internal Server Error"
					}
					c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": errMsg})
					return
				}
			}
		}()
		c.Next()
	}
}

// 获取堆栈信息
func getStack() string {
	buf := make([]byte, 1<<16)
	n := runtime.Stack(buf, false)
	return string(buf[:n])
}
