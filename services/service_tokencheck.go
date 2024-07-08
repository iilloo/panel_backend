package services

import (

	"github.com/gin-gonic/gin"

)

func CheckToken(c *gin.Context) {
	c.Next()
}