package services

import (
	"panel_backend/global"
	"panel_backend/models"
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
	//用户数量减一
	global.UserCount--
	c.JSON(200, gin.H{
		"code": 200,
		"msg":  "删除用户成功",
	})
}

func GetUserName(c *gin.Context) {
	var users []models.User
    var err error
	if users, err = repository.GetUserList(); err != nil {
		c.JSON(500, gin.H{
			"code": 500,
			"msg": "服务器获取用户信息失败",
		})
		return
	}
	userName := users[0].Username
	c.JSON(200, gin.H{
		"code": 200,
		"userName": userName, 
	})

}
type modifyPassword struct {
	NowPassword string `json:"nowPassword" binding:"required"` // 确保字段必须存在
	NewPassword string `json:"newPassword" binding:"required"` // 确保字段必须存在
}
func ModifyPassword(c *gin.Context) {
	var passwordInfo modifyPassword
	c.ShouldBindJSON(&passwordInfo)
	nowpwd := passwordInfo.NowPassword
	newpwd := passwordInfo.NewPassword
	var users []models.User
    var err error
	if users, err = repository.GetUserList(); err != nil {
		c.JSON(500, gin.H{
			"code": 500,
			"msg": "服务器获取用户信息失败",
		})
		return
	}
	if users[0].Password != nowpwd {
		c.JSON(400, gin.H{
			"code": 400,
			"msg": "现有用户密码输入错误",
		})
		return
	}
	user := users[0]
	user.Password = newpwd
	if err := repository.UpdateUser(user); err != nil {
		c.JSON(500, gin.H{
			"code": 500,
			"msg": "服务器更新用户密码时发生错误",
		})
		return
	}
	c.JSON(200, gin.H{
		"code": 200,
		"msg": "更新密码成功",
	})

}