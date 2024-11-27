package v1

import (
	"panel_backend/services"

	"github.com/gin-gonic/gin"
)

func AddTask() gin.HandlerFunc {
	//添加定时任务
	return func(c *gin.Context) {
		services.AddTask(c)
	}
}

func DeleteTask() gin.HandlerFunc {
	//删除定时任务
	return func(c *gin.Context) {
		services.DeleteTask(c)
	}
}

func UpdateTask() gin.HandlerFunc {
	//更新定时任务
	return func(c *gin.Context) {
		services.UpdateTask(c)
	}
}

func GetTaskList() gin.HandlerFunc {
	//获取定时任务列表
	return func(c *gin.Context) {
		services.GetTaskList(c)
	}
}
