package routers

import (
	v1 "panel_backend/api/v1"

	"github.com/gin-gonic/gin"
)



func TimingTaskRouter(router *gin.Engine) {
	//定时任务相关路由
	r := router.Group("/timingTask")
	r.POST("/addTask", v1.AddTask())
	r.POST("/deleteTask", v1.DeleteTask())
	r.POST("/updateTask", v1.UpdateTask())
	r.GET("/getTaskList", v1.GetTaskList())
}