package middlewares

import (
	_"fmt"

	"github.com/gin-gonic/gin"
)

func DeleteRequestMiddle() gin.HandlerFunc {
	return func(c *gin.Context) {
		//删除请求
		if c.GetHeader("Override") == "DELETE" {
			c.Request.Method = "DELETE"
		}
		// for key, value := range c.Request.Header {
        //     fmt.Printf("%s: %s\n", key, value)
        // }
		c.Next()
	}
}