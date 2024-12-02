package handleMiddle

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// Cors 处理跨域问题
func Cors() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD"},
		AllowHeaders:     []string{"Content-Type", "Authorization", "Accept"},
		ExposeHeaders:    []string{"Content-Length", "text/plain", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour, //指定本次预检请求的有效期
	})
}

const XHTTPMethodOverrideHeader = "X-HTTP-Method-Override"

// OverrideMethod 处理方法重写
func OverrideMethod(r *gin.Engine) func(c *gin.Context) {
	return func(c *gin.Context) {
		if c.Request == nil || c.Request.Header == nil {
			return
		}
		// ignores if the request method is not POST
		if c.Request.Method != http.MethodPost {
			return
		}

		method := c.Request.Header.Get(XHTTPMethodOverrideHeader)
		if method == "" {
			return
		}
		// ignores when the overriding method is equal to the request method
		if strings.ToUpper(method) == c.Request.Method {
			return
		}

		switch strings.ToUpper(method) {
		case http.MethodGet:
			c.Request.Method = http.MethodGet
		case http.MethodPost:
			c.Request.Method = http.MethodPost
		case http.MethodPatch:
			c.Request.Method = http.MethodPatch
		case http.MethodPut:
			c.Request.Method = http.MethodPut
		case http.MethodDelete:
			c.Request.Method = http.MethodDelete
		case http.MethodHead:
			c.Request.Method = http.MethodHead
		case http.MethodOptions:
			c.Request.Method = http.MethodOptions
		case http.MethodConnect:
			c.Request.Method = http.MethodConnect
		case http.MethodTrace:
			c.Request.Method = http.MethodTrace
		default:
			// ignore the method value, treated as the original method
			return
		}
		// after the method is overridden, the current request is cancelled and we will need gin.Engine to handle it again
		c.Abort()
		r.HandleContext(c)
	}
}
