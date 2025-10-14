package middleware

import (
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
	"sync"
)

type LocalRateLimiter struct {
	limiterMap map[string]*rate.Limiter
	lock       sync.RWMutex         // 使用RWMutex以优化读取性能
	limit      int                  // 限速器的速率
	burst      int                  // 限速器的临时最大值
	prohibitFn func(c *gin.Context) // 限速器超过限制时的处理函数，非指针类型
}

// NewLocalRateLimiter 创建一个新的LocalRateLimiter实例
func NewLocalRateLimiter(limit int, burst int, prohibitFn func(c *gin.Context)) *LocalRateLimiter {
	limiterMap := make(map[string]*rate.Limiter)

	return &LocalRateLimiter{
		limit:      limit,
		limiterMap: limiterMap,
		burst:      burst,
		prohibitFn: prohibitFn,
	}
}

// HandlerFunc 限速器的中间件
func (l *LocalRateLimiter) HandlerFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method // 当前请求的方法
		path := c.Request.URL.Path // 当前请求的路径

		key := method + ":" + path

		l.lock.RLock()
		limiter, exists := l.limiterMap[key]
		l.lock.RUnlock()

		if !exists {
			l.lock.Lock()
			// 双重检查锁定，以防在获取写锁的过程中limiter被创建
			if limiter, exists = l.limiterMap[key]; !exists {
				limiter = rate.NewLimiter(rate.Limit(l.limit), l.burst)
				l.limiterMap[key] = limiter
			}
			l.lock.Unlock()
		}

		if !limiter.Allow() {
			if l.prohibitFn != nil {
				l.prohibitFn(c)
			} else {
				c.AbortWithStatus(429) // Too Many Requests
			}
			return
		}

		c.Next()
	}
}
