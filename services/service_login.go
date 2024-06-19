package services

import (
	"fmt"
	"panel_backend/global"
	"panel_backend/repository"
	"panel_backend/utils/jwts"
	"time"

	"panel_backend/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func Login(c *gin.Context) {
	//登录
	var user models.User
	c.BindJSON(&user)
	fmt.Printf("user: %v\n", user)
	if global.UserCount == 1 {
		// 查询用户
		repositoryUser, err := repository.GetUserByUsername(user.Username)
		if err != nil {
			global.Log.Error("无此用户", err)
			c.JSON(401, gin.H{
				"code": 401,
				"msg":  "无此用户",
			})
			return
		} else if repositoryUser.Password != user.Password {
			global.Log.Errorf("登录密码错误[%s]", repositoryUser.Password)
			c.JSON(401, gin.H{
				"code": 401,
				"msg":  "登录密码错误",
			})
			return
		}
		claims := jwts.UserClaims{
			Username: user.Username,
			UserID:   1,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(2 * time.Hour)),
			},
		}
		tokenString := jwts.GenerateToken(claims)
		c.Header("panel-token", tokenString)
		global.Log.Debugf("[%s]登录成功\n",user.Username)
		c.JSON(200, gin.H{
			"code": 200,
			"msg":  "登录成功",
		})
	} else if global.UserCount == 0 {
		// 创建用户
		err := repository.CreateUser(user)
		if err != nil {
			global.Log.Error("创建用户失败", err)
			c.JSON(401, gin.H{
				"code": 401,
				"msg":  "创建用户失败",
			})
			return
		}
		claims := jwts.UserClaims{
			Username: user.Username,
			UserID:   1,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * time.Hour)),
			},
		}
		tokenString := jwts.GenerateToken(claims)
		c.Header("panel-token", tokenString)
		c.JSON(200, gin.H{
			"code": 200,
			"msg":  "注册成功",
		})
		global.UserCount = 1
		global.Log.Debugf("[%s]注册成功，当前用户数量: %d\n",user.Username, global.UserCount)
	} else {
		global.Log.Error("用户数量异常", global.UserCount)
		c.JSON(401, gin.H{
			"code": 401,
			"msg":  "系统错误,用户数量异常",
		})
		return
	}
}