package routers

import (
	v1 "panel_backend/api/v1"
	"github.com/gin-gonic/gin"
) 


func WSRouter(router *gin.Engine) {
	router.GET("/ws", v1.WS())
	
}