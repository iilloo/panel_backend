package services

import (
	"panel_backend/global"
	"panel_backend/models"
	"panel_backend/repository"

	"github.com/gin-gonic/gin"
)

func AddTask(c *gin.Context) {
	//添加定时任务
	var timingTask models.TimingTask
	c.ShouldBindJSON(&timingTask)
	global.TaskMu.Lock()
	defer global.TaskMu.Unlock()
	//添加定时任务到定时任务列表
	id, err := global.TaskC.AddFunc(timingTask.Timing, func() {
		//执行定时任务
		global.Log.Infof("定时任务执行：%s", timingTask.TaskName)
	})

	if err != nil {
		global.Log.Errorf("定时任务添加失败：%s", err)
		c.JSON(500, gin.H{
			"code":    500,
			"message": "定时任务添加失败",
		})
		return
	}
	global.TaskMap[timingTask.TaskName] = id
	//添加定时任务到数据库
	err = repository.AddTask(timingTask)
	if err != nil {
		global.Log.Errorf("定时任务添加到数据库失败：%s", err)
		c.JSON(500, gin.H{
			"code":    500,
			"message": "定时任务添加到数据库失败：",
		})
		global.TaskC.Remove(id)
		return
	}
	c.JSON(200, gin.H{
		"code":    200,
		"message": "定时任务添加成功",
	})

}

type DeleteTaskRequest struct {
	TaskNames []string `json:"task_names"`
}
func DeleteTask(c *gin.Context) {
	var timingTask DeleteTaskRequest
	c.ShouldBindJSON(&timingTask)
	global.TaskMu.Lock()
	defer global.TaskMu.Unlock()
	//删除定时任务
	for _, taskName := range timingTask.TaskNames {
		id, ok := global.TaskMap[taskName]
		if !ok {
			global.Log.Errorf("定时任务不存在：%s", taskName)
			c.JSON(500, gin.H{
				"code":    500,
				"message": taskName + "定时任务不存在",
			})
			return
		}

		//删除数据库中的定时任务
		err := repository.DeleteTask(taskName)
		if err != nil {
			global.Log.Errorf("定时任务删除失败：%s", err)
			c.JSON(500, gin.H{
				"code":    500,
				"message": taskName + "定时任务删除失败",
			})
			return
		}
		global.TaskC.Remove(id)
		delete(global.TaskMap, taskName)
	}
	
	c.JSON(200, gin.H{
		"code":    200,
		"message": "定时任务删除成功",
	})
}

type UpdateTaskRequest struct {
	OldTaskName string `json:"old_task_name"`
	NewTaskName string `json:"new_task_name"`
	Timing      string `json:"timing"`
	Command     string `json:"command"`
	Describe    string `json:"describe"`
}

func UpdateTask(c *gin.Context) {
	//更新定时任务
	var updateTaskRequest UpdateTaskRequest
	c.ShouldBindJSON(&updateTaskRequest)
	global.TaskMu.Lock()
	defer global.TaskMu.Unlock()
	//删除旧的定时任务
	id, ok := global.TaskMap[updateTaskRequest.OldTaskName]
	if !ok {
		global.Log.Infof("定时任务不存在：%s", updateTaskRequest.OldTaskName)
		c.JSON(500, gin.H{
			"code":    500,
			"message": "定时任务不存在",
		})
		return
	}
	//删除数据库中的定时任务
	err := repository.DeleteTask(updateTaskRequest.OldTaskName)
	if err != nil {
		global.Log.Errorf("定时任务删除失败：%s", err)
		c.JSON(500, gin.H{
			"code":    500,
			"message": "定时任务删除失败",
		})
		return
	}
	global.TaskC.Remove(id)
	delete(global.TaskMap, updateTaskRequest.OldTaskName)
	//添加新的定时任务
	id, err = global.TaskC.AddFunc(updateTaskRequest.Timing, func() {
		//执行定时任务
		global.Log.Infof("定时任务执行：%s", updateTaskRequest.Command)
	})
	if err != nil {
		global.Log.Errorf("定时任务添加失败：%s", err)
		c.JSON(500, gin.H{
			"code":    500,
			"message": "定时任务添加失败",
		})
		return
	}
	global.TaskMap[updateTaskRequest.NewTaskName] = id
	//添加定时任务到数据库
	err = repository.AddTask(models.TimingTask{
		TaskName: updateTaskRequest.NewTaskName,
		Timing:   updateTaskRequest.Timing,
		Command:  updateTaskRequest.Command,
		Describe: updateTaskRequest.Describe,
	})
	if err != nil {
		global.Log.Errorf("定时任务添加到数据库失败：%s", err)
		c.JSON(500, gin.H{
			"code":    500,
			"message": "定时任务添加到数据库失败",
		})
		global.TaskC.Remove(id)
		return
	}
	c.JSON(200, gin.H{
		"code":    200,
		"message": "定时任务更新成功",
	})
}
func GetTaskList(c *gin.Context) {
	//获取定时任务列表
	tasks, err := repository.GetTaskList()
	if err != nil {
		global.Log.Errorf("获取定时任务列表失败：%s", err)
		c.JSON(500, gin.H{
			"code":    500,
			"message": "获取定时任务列表失败",
		})
		return
	}
	global.Log.Infof("获取定时任务列表成功:%v", tasks)
	c.JSON(200, gin.H{
		"code":    200,
		"tasks":   tasks,
		"message": "获取定时任务列表成功",
	})
}
