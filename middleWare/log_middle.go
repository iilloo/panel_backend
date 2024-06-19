package middlewares

import (
	"fmt"
	"panel_backend/global"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	_"github.com/sirupsen/logrus"
)

//实现gin框架中的自定义日志
func LogMiddle() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		url := c.Request.URL
		ip := strings.Split(c.Request.RemoteAddr, ":")[0]
		
		
		c.Next()
		statusCode := c.Writer.Status()
		method := c.Request.Method
		//求请求响应所花费的总时间
		time := float64(time.Since(startTime).Nanoseconds()) / 1000

		entryStr := fmt.Sprintf("\033[34m[GIN]\033[0m %s |%d| %10.3fµs | %s | %s | \"%s\"\n",
			startTime.Format("2006-01-02 15:04:05"), statusCode, time, ip, method, url)

		global.Log.Info(entryStr)
	}

}
