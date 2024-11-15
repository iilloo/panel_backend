package services

import (
	"panel_backend/global"
	"panel_backend/repository"

	"github.com/gin-gonic/gin"
)
type DeleteAccountRequest struct {
	Username string `json:"username" binding:"required"` // 确保字段必须存在
}
func DeleteUser(c *gin.Context) {
	var request DeleteAccountRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(400, gin.H{
			"code": 400,
			"msg":  "请求参数错误",
		})
		return
	}
	userName := request.Username
	global.Log.Infof("删除用户[%s]\n", userName)
	//删除用户
	if err := repository.DeleteUserByName(userName); err != nil {
		c.JSON(500, gin.H{
			"code": 500,
			"msg":  "删除用户失败",
		})
		return
	}
	c.JSON(200, gin.H{
		"code": 200,
		"msg":  "删除用户成功",
	})
}