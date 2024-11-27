package main

import (
	_ "fmt"
	_ "net/http"
	"panel_backend/core"
	"panel_backend/global"
	"panel_backend/routers"
	_ "panel_backend/utils/jwts"
	_ "time"

	"panel_backend/tests"

	_ "github.com/gin-gonic/gin"
	_ "github.com/golang-jwt/jwt/v5"

	_ "github.com/sirupsen/logrus"
	_"panel_backend/models"
)

func main() {
	
	// 初始化自定义日志
	global.Log = core.InitLogger("Log_panel/", "panel")
	// 初始化配置项
	global.Config = core.InitConfig()
	// 初始化数据库
	global.DB = core.InitMysql()
	global.UserCount = core.InitUserCount()
	// 初始化JWT
	global.Secret = core.InitJwt()
	// 初始化bash
	if bash := core.InitCMD(); bash != nil {
		global.Bash = bash
	} else {
		global.Log.Errorf("bash初始化失败")
		return
	}
	// global.CMD = exec.Command("bash")
	// // 自动迁移创建表
	
	// err := global.DB.AutoMigrate(&models.TimingTask{})
	// if err != nil {
	// 	panic("Failed to migrate TimingTask table: " + err.Error())
	// }

	// 初始化路由
	router := routers.Routers()

	// 测试路由
	router.GET("/hello", tests.RouteHello())
	// tests.CreateUser()
	// tests.CreateUserTest()
	// tests.DeleteUser("kazusa")

	// 从配置项中获得地址并启动服务
	addr := global.Config.System.GetAddr()
	router.Run(addr)
}
