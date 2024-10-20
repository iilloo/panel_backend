package middlewares

import (
	"fmt"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// 使用cors解决浏览器跨域问题
func CorsMiddle() gin.HandlerFunc {

	config := cors.DefaultConfig()

	config.AllowOriginFunc = func(origin string) bool {
		return true
	}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "HEAD"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Authorization", "Accept", "token", "Override"}
	config.ExposeHeaders = []string{"Content-Length", "panel-token", "Need-ResponseHeader", "Content-Disposition"}
	config.AllowCredentials = true
	config.MaxAge = time.Hour

	fmt.Println("cors中间件加载成功")
	return cors.New(config)
}
