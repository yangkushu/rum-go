package iface

import "github.com/gin-gonic/gin"

type IMiddleware interface {
	HandlerFunc() gin.HandlerFunc
}
