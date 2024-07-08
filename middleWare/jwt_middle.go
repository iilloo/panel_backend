package middlewares

import (
	"net/http"
	"panel_backend/global"
	"panel_backend/utils/jwts"
	"strings"

	"github.com/gin-gonic/gin"
)

func JWTMiddle() gin.HandlerFunc {
	return func(c *gin.Context) {
		//排除登录接口和websocket接口
		if (c.Request.URL.Path == "/login" || c.Request.URL.Path == "/ws") {
			c.Next()
		} else {
			authoToken := c.GetHeader("Authorization")
			if authoToken == "" {
				global.Log.Warnf("[%s]请求未携带token，无权限访问\n",c.RemoteIP())
				c.JSON(http.StatusUnauthorized, gin.H{
					"code": 401,
					"msg":  "请求未携带token，无权限访问",
				})
				c.Abort()
				return
			}
			//分割一下
			segs := strings.Split(authoToken, " ")
			if len(segs) != 2 || segs[0] != "Bearer" {
				global.Log.Warnf("[%s]请求携带非法token，无权限访问\n",c.RemoteIP())
				c.JSON(http.StatusUnauthorized, gin.H{
					"code": 401,
					"msg":  "请求携带非法token，无权限访问",
				})
				c.Abort()
				return
			}
			tokenString := segs[1]
			//解析并刷新token
			claims, newTokenString, err := jwts.ParseToken(tokenString)
			if err != nil {
				global.Log.Warnf("[%s]token无效,解析失败\n",c.RemoteIP())
				c.JSON(http.StatusUnauthorized, gin.H{
					"code": 401,
					"msg":  "token无效,解析失败",
				})
				c.Abort()
				return
			}
			//设置刷新后的token
			if newTokenString != tokenString {
				c.Header("panel-token", newTokenString)
			}
			//中间件传递参数
			c.Set("Username", claims.Username)
			c.Set("UserID", claims.UserID)
			c.Next()
		}
	}
	
}