package v1

import (
	"github.com/gin-gonic/gin"
	"panel_backend/services"
)

func HostBasicInfos() gin.HandlerFunc{
	//获取主机基本信息
	return func(c *gin.Context) {
		services.HostBasicInfos(c)
	}
}