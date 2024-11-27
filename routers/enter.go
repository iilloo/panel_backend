package routers

import (

	"panel_backend/middleWare"

	"github.com/gin-gonic/gin"
)

func Routers() *gin.Engine{
	router := gin.New()
	// 一些中间件
	// router.Use(func(c *gin.Context) {
	// 	if c.Request.Method == "OPTIONS" {
	// 		fmt.Println("options@@@@@@@@@@@@@@@@@@@")
	// 		c.AbortWithStatus(http.StatusOK)
	// 		return
	// 	}
	// 	c.Next()
	// })
	//使用cors中间件，解决浏览器跨域问题
	router.Use(middlewares.CorsMiddle())
	// 使用删除请求中间件
	router.Use(middlewares.DeleteRequestMiddle())
	
	// 使用jwt中间件
	router.Use(middlewares.JWTMiddle())
	// 使用日志中间件
	router.Use(middlewares.LogMiddle())
	
	//登录相关路由
	LoginRouter(router)

	//文件系统相关路由
	FileSysRouter(router)

	//websocket
	WSRouter(router)

	//主机状态相关路由
	HostStatusRouter(router)
	//token检查相关路由
	CheckToken(router)
	//用户管理相关路由
	UserManageRouter(router)
	//定时任务相关路由
	TimingTaskRouter(router)


	return router
}