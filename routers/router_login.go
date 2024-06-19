package routers

import (
	v1 "panel_backend/api/v1"

	"github.com/gin-gonic/gin"
)

func LoginRouter(router *gin.Engine) {
	//登录相关路由

	router.POST("/login", v1.Login())

}
