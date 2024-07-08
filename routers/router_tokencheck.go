package routers

import (
	v1 "panel_backend/api/v1"

	"github.com/gin-gonic/gin"
)

func CheckToken(router *gin.Engine) {
	router.GET("/checkToken", v1.CheckToken())
}