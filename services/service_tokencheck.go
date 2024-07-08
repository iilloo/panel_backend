package services

import (
	"panel_backend/global"

	"github.com/gin-gonic/gin"
)

func CheckToken(c *gin.Context) {
	c.JSON(200, gin.H{
		"code": 200,
		"msg":  "token check success",
	})
	global.Log.Infof("token check success")
	c.Next()
}
