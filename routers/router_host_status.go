package routers

import (
	v1 "panel_backend/api/v1"
	"github.com/gin-gonic/gin"
)

func HostStatusRouter(router *gin.Engine) {
	r := router.Group("/hostStatus")
	r.GET("/hostBasicInfo", v1.HostBasicInfos())
	
}