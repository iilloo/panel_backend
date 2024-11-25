package v1

import (
	"panel_backend/services"

	"github.com/gin-gonic/gin"
)

func CreateUser(c *gin.Context) {
	//创建用户
}

func UpdateUser(c *gin.Context) {
	//更新用户
}

func DeleteUser(c *gin.Context) {
	//删除用户
	services.DeleteUser(c)
}

func GetUserName(c *gin.Context) {
	//获取用户
	services.GetUserName(c)
}
func ModifyPassword(c *gin.Context) {
	//修改唯一用户的密码
	services.ModifyPassword(c)
}