package routers

import (
	v1 "panel_backend/api/v1"

	"github.com/gin-gonic/gin"
)

func UserManageRouter(router *gin.Engine) {
	//用户管理相关路由
	// router.POST("/updateUser", v1.UpdateUser)
	router.POST("/deleteAccount", v1.DeleteUser)

	router.GET("/getUserName", v1.GetUserName)
	router.POST("/modifyPassword", v1.ModifyPassword)
}