package v1

import (
	"panel_backend/services"

	"github.com/gin-gonic/gin"
)

func CheckToken() gin.HandlerFunc{
	return func(c *gin.Context) {
		//检查token是否有效
		services.CheckToken(c)
	}
}